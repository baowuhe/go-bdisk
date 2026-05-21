package bdisk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// DownloadProgress 下载进度信息
type DownloadProgress struct {
	Downloaded int64   `json:"downloaded"` // 已下载字节数
	Total      int64   `json:"total"`      // 总字节数
	Percent    float64 `json:"percent"`    // 下载百分比 (0-100)
}

// ProgressCallback 进度回调函数类型
type ProgressCallback func(progress DownloadProgress)

// DownloadService 下载服务
type DownloadService struct {
	client *Client
}

// Start 开始下载文件
func (d *DownloadService) Start(remotePath, localPath string) error {
	return d.StartWithProgress(remotePath, localPath, nil)
}

// StartWithProgress 开始下载文件，带进度回调
func (d *DownloadService) StartWithProgress(remotePath, localPath string, callback ProgressCallback) error {
	if d.client.token == nil || !d.client.token.IsValid() {
		return ErrTokenExpired
	}

	// 步骤 1: 获取文件信息和 fs_id
	fsID, fileSize, fileName, err := d.getFileFSID(remotePath)
	if err != nil {
		return err
	}

	// 步骤 2: 使用 filemetas 接口（multimedia 路径）获取下载链接 dlink
	dlink, err := d.getDlinkFromMultimedia(fsID, remotePath)
	if err != nil {
		return err
	}

	// 步骤 3: 使用 dlink 下载文件
	return d.downloadFileOfficial(dlink, localPath, fileName, fileSize, callback)
}

// GetFileFSID 获取文件的 fs_id、文件大小和文件名
func (d *DownloadService) GetFileFSID(path string) (int64, int64, string, error) {
	return d.getFileFSID(path)
}

// getFileFSID 获取文件的 fs_id、文件大小和文件名
func (d *DownloadService) getFileFSID(path string) (int64, int64, string, error) {
	var result struct {
		List []struct {
			FSID           int64  `json:"fs_id"`
			Isdir          int    `json:"isdir"`
			Size           int64  `json:"size"`
			ServerFilename string `json:"server_filename"`
		} `json:"list"`
	}

	params := url.Values{
		"method": {"meta"},
		"path":   {path},
	}

	err := d.client.doRequest("GET", "/xpan/file", params, &result)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to get file info: %w", err)
	}

	if len(result.List) == 0 {
		return 0, 0, "", fmt.Errorf("file not found: %s", path)
	}

	if result.List[0].Isdir == 1 {
		return 0, 0, "", fmt.Errorf("cannot download directory: %s", path)
	}

	return result.List[0].FSID, result.List[0].Size, result.List[0].ServerFilename, nil
}

// getDlinkFromMultimedia 使用 /xpan/multimedia 路径获取 dlink
func (d *DownloadService) getDlinkFromMultimedia(fsID int64, path string) (string, error) {
	// 构建请求 URL - 官方使用 /xpan/multimedia 路径
	reqURL := fmt.Sprintf("%s/xpan/multimedia?method=filemetas", OpenAPIBaseURL)

	params := url.Values{}
	params.Set("access_token", d.client.token.AccessToken)

	// 构建 fsids 数组
	fsids := []int64{fsID}
	fsidsJSON, err := json.Marshal(fsids)
	if err != nil {
		return "", err
	}
	params.Set("fsids", string(fsidsJSON))
	params.Set("dlink", "1")       // 必选参数，才能拿到 dlink
	params.Set("path", path)       // 查询共享目录或专属空间内文件时需要
	params.Set("thumb", "1")       // 官方 demo 中的参数
	params.Set("needmedia", "1")   // 官方 demo 中的参数
	params.Set("extra", "1")       // 官方 demo 中的参数

	reqURL = reqURL + "&" + params.Encode()

	// 创建请求（使用 POST 方法，官方 demo 使用 POST）
	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Host", "pan.baidu.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := d.client.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 限制响应体大小为 10MB
	limitedReader := io.LimitReader(resp.Body, 10*1024*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	// 解析响应
	var result struct {
		Errno int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		List  []struct {
			Dlink string `json:"dlink"`
		} `json:"list"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if result.Errno != 0 {
		return "", fmt.Errorf("filemetas error: %s (errno: %d)", result.Errmsg, result.Errno)
	}

	if len(result.List) == 0 || result.List[0].Dlink == "" {
		return "", fmt.Errorf("no dlink found in response: %s", string(body))
	}

	return result.List[0].Dlink, nil
}

// downloadFileOfficial 使用官方 demo 的方式下载文件
func (d *DownloadService) downloadFileOfficial(dlink, localPath, fileName string, fileSize int64, callback ProgressCallback) error {
	// 如果 localPath 是文件夹（包括 "." 这种情况），使用网盘文件名
	if fileInfo, err := os.Stat(localPath); err == nil && fileInfo.IsDir() {
		localPath = filepath.Join(localPath, fileName)
	}

	// 构建下载 URL：dlink + "&access_token=" + access_token
	downloadURL := dlink + "&access_token=" + d.client.token.AccessToken

	// 创建 POST 请求（官方 demo 使用 POST）
	req, err := http.NewRequest("POST", downloadURL, nil)
	if err != nil {
		return err
	}

	// 设置 User-Agent（官方要求）
	req.Header.Set("User-Agent", "pan.baidu.com")

	// 发送请求，使用客户端的 http.Client（已配置好 TLS）
	resp, err := d.client.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed, status: %d, body: %s", resp.StatusCode, string(body))
	}

	// 确保目录存在
	dir := filepath.Dir(localPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// 创建文件
	file, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// 如果需要进度回调，使用 io.TeeReader 包装
	if callback != nil && fileSize > 0 {
		var downloaded int64 = 0
		reader := &progressReader{
			reader: resp.Body,
			onRead: func(n int64) {
				downloaded += n
				percent := float64(downloaded) / float64(fileSize) * 100
				callback(DownloadProgress{
					Downloaded: downloaded,
					Total:      fileSize,
					Percent:    percent,
				})
			},
		}

		_, err = io.Copy(file, reader)
		return err
	}

	// 直接复制响应体到文件（支持大文件，不全部读入内存）
	_, err = io.Copy(file, resp.Body)
	return err
}

// progressReader 用于跟踪读取进度的 Reader
type progressReader struct {
	reader io.Reader
	onRead func(int64)
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 && r.onRead != nil {
		r.onRead(int64(n))
	}
	return n, err
}
