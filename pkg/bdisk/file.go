package bdisk

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/baowuhe/go-bdisk/pkg/bdisk/model"
)

// FileListInfo 文件列表信息（用于 filemanager 接口）
type FileListInfo struct {
	Path    string `json:"path"`    // 源文件路径
	Newname string `json:"newname"` // 新文件名（rename 时使用）
	Dest    string `json:"dest"`    // 目标路径（copy/move 时使用）
	Ondup   string `json:"ondup"`   // 重名处理策略：fail-失败，newcopy-覆盖，skip-跳过
}

// FileManagerReturn filemanager 接口返回
type FileManagerReturn struct {
	Errno     int                      `json:"errno"`
	Info      []map[string]interface{} `json:"info"`
	Taskid    int                      `json:"taskid"` // 异步操作时返回
	RequestID int                      `json:"request_id"`
}

// fileManager 调用文件管理接口
func (f *FileService) fileManager(opera string, async string, fileList []FileListInfo) (*FileManagerReturn, error) {
	uri := "http://pan.baidu.com/rest/2.0/xpan/file?method=filemanager&"

	params := url.Values{}
	params.Set("access_token", f.client.token.AccessToken)
	params.Set("opera", opera)
	uri += params.Encode()

	headers := map[string]string{
		"Host":         "pan.baidu.com",
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// 构建请求体
	postBody := url.Values{}
	postBody.Add("async", async)
	fileListJSON, err := json.Marshal(fileList)
	if err != nil {
		return nil, err
	}
	postBody.Add("filelist", string(fileListJSON))

	body, _, err := f.client.doHTTPRequest(uri, strings.NewReader(postBody.Encode()), headers)
	if err != nil {
		return nil, err
	}

	var ret FileManagerReturn
	if err := json.Unmarshal([]byte(body), &ret); err != nil {
		return nil, fmt.Errorf("unmarshal filemanager body failed: %w, body: %s", err, body)
	}

	if ret.Errno != 0 {
		return nil, fmt.Errorf("filemanager %s failed, errno: %d", opera, ret.Errno)
	}

	return &ret, nil
}

// List 列出文件
func (f *FileService) List(path string) (*model.FileList, error) {
	var result struct {
		List []struct {
			FSID           int64  `json:"fs_id"`
			Path           string `json:"path"`
			ServerFilename string `json:"server_filename"`
			Isdir          int    `json:"isdir"`
			Size           int64  `json:"size"`
			ServerMtime    int64  `json:"server_mtime"`
			ServerCtime    int64  `json:"server_ctime"`
			MD5            string `json:"md5"`
		} `json:"list"`
	}

	params := url.Values{
		"method": {"list"},
		"dir":    {path},
		"order":  {"time"},
		"desc":   {"1"},
	}

	err := f.client.doRequest("GET", "/xpan/file", params, &result)
	if err != nil {
		return nil, err
	}

	items := make([]*model.FileInfo, 0, len(result.List))
	for _, item := range result.List {
		fileType := model.FileTypeFile
		if item.Isdir == 1 {
			fileType = model.FileTypeDir
		}

		items = append(items, &model.FileInfo{
			FSID:       item.FSID,
			Path:       item.Path,
			Name:       item.ServerFilename,
			Type:       fileType,
			Size:       item.Size,
			ModifiedAt: time.Unix(item.ServerMtime, 0),
			CreatedAt:  time.Unix(item.ServerCtime, 0),
			MD5:        item.MD5,
		})
	}

	return &model.FileList{
		Path:  path,
		Items: items,
		Total: len(items),
	}, nil
}

// GetInfo 获取文件信息
func (f *FileService) GetInfo(path string) (*model.FileInfo, error) {
	var result struct {
		List []struct {
			FSID           int64  `json:"fs_id"`
			Path           string `json:"path"`
			ServerFilename string `json:"server_filename"`
			Isdir          int    `json:"isdir"`
			Size           int64  `json:"size"`
			ServerMtime    int64  `json:"server_mtime"`
			ServerCtime    int64  `json:"server_ctime"`
			MD5            string `json:"md5"`
		} `json:"list"`
	}

	params := url.Values{
		"method": {"meta"},
		"path":   {path},
	}

	err := f.client.doRequest("GET", "/xpan/file", params, &result)
	if err != nil {
		return nil, err
	}

	if len(result.List) == 0 {
		return nil, fmt.Errorf("file not found")
	}

	item := result.List[0]
	fileType := model.FileTypeFile
	if item.Isdir == 1 {
		fileType = model.FileTypeDir
	}

	return &model.FileInfo{
		FSID:       item.FSID,
		Path:       item.Path,
		Name:       item.ServerFilename,
		Type:       fileType,
		Size:       item.Size,
		ModifiedAt: time.Unix(item.ServerMtime, 0),
		CreatedAt:  time.Unix(item.ServerCtime, 0),
		MD5:        item.MD5,
	}, nil
}

// Copy 复制文件/文件夹
// srcPath: 源路径
// destPath: 目标路径（文件夹）
// ondup: 重名处理策略，可选值：fail(默认，失败), newcopy(覆盖), skip(跳过)
func (f *FileService) Copy(srcPath, destPath string, ondup ...string) error {
	ondupValue := "fail"
	if len(ondup) > 0 {
		ondupValue = ondup[0]
	}

	fileList := []FileListInfo{
		{
			Path:  srcPath,
			Dest:  destPath,
			Ondup: ondupValue,
		},
	}

	_, err := f.fileManager("copy", "2", fileList)
	return err
}

// Move 移动文件/文件夹
// srcPath: 源路径
// destPath: 目标路径（文件夹）
// ondup: 重名处理策略，可选值：fail(默认，失败), newcopy(覆盖), skip(跳过)
func (f *FileService) Move(srcPath, destPath string, ondup ...string) error {
	ondupValue := "fail"
	if len(ondup) > 0 {
		ondupValue = ondup[0]
	}

	fileList := []FileListInfo{
		{
			Path:  srcPath,
			Dest:  destPath,
			Ondup: ondupValue,
		},
	}

	_, err := f.fileManager("move", "2", fileList)
	return err
}

// Delete 删除文件/文件夹
// paths: 要删除的文件或文件夹路径列表
func (f *FileService) Delete(paths ...string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no paths provided")
	}

	fileList := make([]FileListInfo, 0, len(paths))
	for _, path := range paths {
		fileList = append(fileList, FileListInfo{
			Path: path,
		})
	}

	_, err := f.fileManager("delete", "2", fileList)
	return err
}

// Rename 重命名文件/文件夹
// path: 文件或文件夹的当前路径
// newName: 新的名称（不包含路径）
func (f *FileService) Rename(path, newName string) error {
	fileList := []FileListInfo{
		{
			Path:    path,
			Newname: newName,
		},
	}

	_, err := f.fileManager("rename", "2", fileList)
	return err
}

// CreateDirReturn 创建文件夹返回结果
type CreateDirReturn struct {
	Errno    int    `json:"errno"`
	FSID     uint64 `json:"fs_id"`
	Category int    `json:"category"`
	Path     string `json:"path"`
	Ctime    uint64 `json:"ctime"`
	Mtime    uint64 `json:"mtime"`
	Isdir    int    `json:"isdir"`
}

// CreateDir 创建文件夹
// path: 要创建的文件夹绝对路径
// rtype: 文件命名策略，0-不重命名返回冲突，1-冲突即重命名（默认）
func (f *FileService) CreateDir(path string, rtype ...int) (*CreateDirReturn, error) {
	rtypeValue := 1
	if len(rtype) > 0 {
		rtypeValue = rtype[0]
	}

	uri := "https://pan.baidu.com/rest/2.0/xpan/file?method=create&"

	params := url.Values{}
	params.Set("access_token", f.client.token.AccessToken)
	uri += params.Encode()

	headers := map[string]string{
		"Host":         "pan.baidu.com",
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// 构建请求体
	postBody := url.Values{}
	postBody.Add("path", path)
	postBody.Add("isdir", "1")
	postBody.Add("rtype", fmt.Sprintf("%d", rtypeValue))

	body, _, err := f.client.doHTTPRequest(uri, strings.NewReader(postBody.Encode()), headers)
	if err != nil {
		return nil, err
	}

	var ret CreateDirReturn
	if err := json.Unmarshal([]byte(body), &ret); err != nil {
		return nil, fmt.Errorf("unmarshal create dir body failed: %w, body: %s", err, body)
	}

	if ret.Errno != 0 {
		return nil, fmt.Errorf("create dir failed, errno: %d", ret.Errno)
	}

	return &ret, nil
}
