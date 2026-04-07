package bdisk

import (
	"crypto/tls"
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
	timeout := 60 * time.Second
	retryTimes := 3

	var resp *http.Response
	var err error

	for i := 1; i <= retryTimes; i++ {
		// 如果是 *strings.Reader 或 *bytes.Reader，需要重置位置
		if seeker, ok := body.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		tr := &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConnsPerHost: -1,
		}
		httpClient := &http.Client{
			Transport: tr,
			Timeout:   timeout,
		}

		req, err := http.NewRequest("POST", urlStr, body)
		if err != nil {
			return "", 0, err
		}

		// request header
		for k, v := range headers {
			req.Header.Add(k, v)
		}

		resp, err = httpClient.Do(req)
		if err == nil {
			break
		}
		if i == retryTimes {
			return "", 0, err
		}
		// 等待一小段时间后重试
		time.Sleep(time.Second)
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
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
