package bdisk

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/baowuhe/go-bdisk/pkg/bdisk/model"
)

// 分片大小：官方 demo 使用 4MB
const UploadBlockSize = 4 * 1024 * 1024

// UploadProgress 上传进度信息
type UploadProgress struct {
	Uploaded      int64   `json:"uploaded"`      // 已上传字节数
	Total         int64   `json:"total"`         // 总字节数
	Percent       float64 `json:"percent"`       // 上传百分比 (0-100)
	CurrentPart   int     `json:"current_part"`  // 当前分片序号
	TotalParts    int     `json:"total_parts"`   // 总分片数
}

// UploadProgressCallback 上传进度回调函数类型
type UploadProgressCallback func(progress UploadProgress)

// UploadService 上传服务
type UploadService struct {
	client *Client
}

// Start 开始上传文件
func (u *UploadService) Start(localPath, remotePath string) (string, error) {
	return u.StartWithProgress(localPath, remotePath, nil)
}

// StartWithProgress 开始上传文件，带进度回调，返回实际使用的远程路径
func (u *UploadService) StartWithProgress(localPath, remotePath string, callback UploadProgressCallback) (string, error) {
	if u.client.token == nil || !u.client.token.IsValid() {
		return "", ErrTokenExpired
	}

	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("stat file failed: %w", err)
	}

	fileSize := fileInfo.Size()

	// 获取本地文件名
	localFileName := filepath.Base(localPath)

	// 检查目标路径是否为文件夹，如果是则拼接文件名
	destInfo, err := u.client.File.GetInfo(remotePath)
	if err == nil && destInfo.Type == model.FileTypeDir {
		// 目标为文件夹，将本地文件名拼接到目标路径
		remotePath = strings.TrimSuffix(remotePath, "/") + "/" + localFileName
	} else if err != nil {
		// 目标路径不存在或其他错误，不拼接，直接使用原路径
		// 只有目标是已存在的文件夹时才拼接文件名
	}

	// 计算文件分片的 MD5 列表
	blockList, err := u.calculateBlockMD5(file, fileSize)
	if err != nil {
		return "", fmt.Errorf("calculate block MD5 failed: %w", err)
	}

	// 步骤 1: precreate - 预上传
	precreateRet, err := u.precreate(remotePath, fileSize, blockList)
	if err != nil {
		return "", fmt.Errorf("precreate failed: %w", err)
	}

	// 步骤 2: upload - 分片上传
	if err := u.uploadParts(file, fileSize, remotePath, precreateRet.Uploadid, callback); err != nil {
		return "", fmt.Errorf("upload parts failed: %w", err)
	}

	// 步骤 3: create - 创建文件
	if _, err := u.create(remotePath, fileSize, blockList, precreateRet.Uploadid); err != nil {
		return "", fmt.Errorf("create failed: %w", err)
	}

	return remotePath, nil
}

// calculateBlockMD5 计算每个分片的 MD5
func (u *UploadService) calculateBlockMD5(file *os.File, fileSize int64) ([]string, error) {
	var blockList []string

	// 重置文件指针
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, UploadBlockSize)
	for offset := int64(0); offset < fileSize; {
		// 读取分片
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		// 计算 MD5
		hash := md5.Sum(buf[:n])
		md5Str := hex.EncodeToString(hash[:])
		blockList = append(blockList, md5Str)

		offset += int64(n)
	}

	// 重置文件指针
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return blockList, nil
}

// precreateArg precreate 接口参数
type precreateArg struct {
	Path      string   `json:"path"`
	Size      uint64   `json:"size"`
	BlockList []string `json:"block_list"`
	Isdir     string   `json:"isdir"`
	Autoinit  string   `json:"autoinit"`
	Rtype     string   `json:"rtype"`
}

// precreateReturn precreate 接口返回
type precreateReturn struct {
	Errno      int    `json:"errno"`
	ReturnType int    `json:"return_type"`
	BlockList  []int  `json:"block_list"`
	Uploadid   string `json:"uploadid"`
	RequestID  int    `json:"request_id"`
}

// precreate 预上传，获取 uploadid
func (u *UploadService) precreate(path string, size int64, blockList []string) (*precreateReturn, error) {
	uri := "https://pan.baidu.com/rest/2.0/xpan/file?method=precreate&"

	params := url.Values{}
	params.Set("access_token", u.client.token.AccessToken)
	uri += params.Encode()

	headers := map[string]string{
		"Host":         "pan.baidu.com",
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// 构建请求体
	blockListJSON, _ := json.Marshal(blockList)
	postBody := url.Values{}
	postBody.Add("path", path)
	postBody.Add("size", strconv.FormatUint(uint64(size), 10))
	postBody.Add("block_list", string(blockListJSON))
	postBody.Add("isdir", "0")
	postBody.Add("autoinit", "1")
	// 当 path 冲突且 block_list 不同时，进行重命名
	postBody.Add("rtype", "2")

	body, _, err := u.client.doHTTPRequest(uri, strings.NewReader(postBody.Encode()), headers)
	if err != nil {
		return nil, err
	}

	var ret precreateReturn
	if err := json.Unmarshal([]byte(body), &ret); err != nil {
		return nil, fmt.Errorf("unmarshal precreate body failed: %w, body: %s", err, body)
	}

	if ret.Errno != 0 {
		return nil, fmt.Errorf("precreate failed, errno: %d", ret.Errno)
	}

	return &ret, nil
}

// uploadParts 上传所有分片
func (u *UploadService) uploadParts(file *os.File, fileSize int64, path string, uploadid string, callback UploadProgressCallback) error {
	totalParts := int((fileSize + UploadBlockSize - 1) / UploadBlockSize)
	var uploaded int64 = 0

	for partseq := 0; partseq < totalParts; partseq++ {
		// 计算当前分片的偏移量
		offset := int64(partseq) * UploadBlockSize

		// 读取分片数据
		buf := make([]byte, UploadBlockSize)
		n, err := file.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// 上传分片
		if err := u.uploadPart(path, uploadid, partseq, buf[:n]); err != nil {
			return err
		}

		uploaded += int64(n)

		// 回调进度
		if callback != nil {
			percent := float64(uploaded) / float64(fileSize) * 100
			callback(UploadProgress{
				Uploaded:    uploaded,
				Total:       fileSize,
				Percent:     percent,
				CurrentPart: partseq + 1,
				TotalParts:  totalParts,
			})
		}
	}

	return nil
}

// uploadPart 上传单个分片
func (u *UploadService) uploadPart(path string, uploadid string, partseq int, data []byte) error {
	uri := "https://d.pcs.baidu.com/rest/2.0/pcs/superfile2?method=upload&"

	params := url.Values{}
	params.Set("access_token", u.client.token.AccessToken)
	params.Set("path", path)
	params.Set("uploadid", uploadid)
	params.Set("partseq", strconv.Itoa(partseq))
	uri += params.Encode()

	headers := map[string]string{
		"Host": "d.pcs.baidu.com",
	}

	// 构建 multipart 请求
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("file", "file")
	if err != nil {
		return err
	}

	// 写入数据
	_, err = fileWriter.Write(data)
	if err != nil {
		return err
	}

	bodyWriter.Close()

	headers["Content-Type"] = bodyWriter.FormDataContentType()

	body, _, err := u.client.doHTTPRequest(uri, bodyBuf, headers)
	if err != nil {
		return err
	}

	// 解析响应
	var ret struct {
		Md5       string `json:"md5"`
		RequestID int    `json:"request_id"`
	}
	if err := json.Unmarshal([]byte(body), &ret); err != nil {
		return fmt.Errorf("unmarshal upload body failed: %w, body: %s", err, body)
	}

	if ret.Md5 == "" {
		return fmt.Errorf("upload failed: md5 is empty, body: %s", body)
	}

	return nil
}

// createArg create 接口参数
type createArg struct {
	Uploadid  string   `json:"uploadid"`
	Path      string   `json:"path"`
	Size      uint64   `json:"size"`
	Isdir     string   `json:"isdir"`
	BlockList []string `json:"block_list"`
	Rtype     string   `json:"rtype"`
}

// createReturn create 接口返回
type createReturn struct {
	Errno int    `json:"errno"`
	Path  string `json:"path"`
}

// create 创建文件，完成上传
func (u *UploadService) create(path string, size int64, blockList []string, uploadid string) (*createReturn, error) {
	uri := "https://pan.baidu.com/rest/2.0/xpan/file?method=create&"

	params := url.Values{}
	params.Set("access_token", u.client.token.AccessToken)
	uri += params.Encode()

	headers := map[string]string{
		"Host":         "pan.baidu.com",
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// 构建请求体
	blockListJSON, _ := json.Marshal(blockList)
	postBody := url.Values{}
	postBody.Add("rtype", "2")
	postBody.Add("path", path)
	postBody.Add("size", strconv.FormatUint(uint64(size), 10))
	postBody.Add("isdir", "0")
	postBody.Add("block_list", string(blockListJSON))
	postBody.Add("uploadid", uploadid)

	body, _, err := u.client.doHTTPRequest(uri, strings.NewReader(postBody.Encode()), headers)
	if err != nil {
		return nil, err
	}

	var ret createReturn
	if err := json.Unmarshal([]byte(body), &ret); err != nil {
		return nil, fmt.Errorf("unmarshal create body failed: %w, body: %s", err, body)
	}

	if ret.Errno != 0 {
		return nil, fmt.Errorf("create failed, errno: %d", ret.Errno)
	}

	return &ret, nil
}
