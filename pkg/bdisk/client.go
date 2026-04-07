package bdisk

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 百度网盘 SDK 客户端
type Client struct {
	config *Config
	token  *Token
	http   *http.Client

	// 子模块
	Auth     *AuthService
	User     *UserService
	File     *FileService
	Download *DownloadService
	Upload   *UploadService
}

// NewClient 创建新的 SDK 客户端
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 创建 HTTP Transport 和 Client，设置合理超时
	transport := &http.Transport{
		MaxIdleConnsPerHost: 10,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Minute,
	}

	c := &Client{
		config: config,
		http:   httpClient,
	}

	// 初始化子服务
	c.Auth = &AuthService{client: c}
	c.User = &UserService{client: c}
	c.File = &FileService{client: c}
	c.Download = &DownloadService{client: c}
	c.Upload = &UploadService{client: c}

	return c, nil
}

// SetToken 设置访问令牌
func (c *Client) SetToken(token *Token) {
	c.token = token
}

// GetToken 获取当前访问令牌
func (c *Client) GetToken() *Token {
	return c.token
}

// ClearToken 清除访问令牌
func (c *Client) ClearToken() {
	c.token = nil
}

// doHTTPRequest 发送 HTTP 请求（参考官方 demo）
func (c *Client) doHTTPRequest(urlStr string, body io.Reader, headers map[string]string) (string, int, error) {
	retryTimes := 3

	var resp *http.Response
	var err error

	for i := 1; i <= retryTimes; i++ {
		// 如果是 *strings.Reader 或 *bytes.Reader，需要重置位置
		if seeker, ok := body.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		req, err := http.NewRequest("POST", urlStr, body)
		if err != nil {
			return "", 0, err
		}

		// request header
		for k, v := range headers {
			req.Header.Add(k, v)
		}

		resp, err = c.http.Do(req)
		if err == nil {
			break
		}
		if i == retryTimes {
			return "", 0, err
		}
		// 等待一小段时间后重试
		time.Sleep(time.Second)
	}

	// resp 可能为 nil（所有重试都失败）
	if resp == nil {
		return "", 0, fmt.Errorf("request failed after %d retries", retryTimes)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", resp.StatusCode, fmt.Errorf("HTTP error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	// 限制响应体大小为 100MB，防止恶意服务器发送大量数据
	limitedReader := io.LimitReader(resp.Body, 100*1024*1024)
	respBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", resp.StatusCode, err
	}

	return string(respBody), resp.StatusCode, nil
}

// AuthService 认证服务
type AuthService struct {
	client *Client
}

// UserService 用户服务
type UserService struct {
	client *Client
}

// FileService 文件服务
type FileService struct {
	client *Client
}
