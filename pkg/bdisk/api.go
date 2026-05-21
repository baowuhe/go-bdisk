package bdisk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// doRequest 执行API请求
func (c *Client) doRequest(method, endpoint string, params url.Values, result interface{}) error {
	if c.token == nil || !c.token.IsValid() {
		return ErrTokenExpired
	}

	if params == nil {
		params = url.Values{}
	}
	params.Set("access_token", c.token.AccessToken)

	reqURL := fmt.Sprintf("%s%s?%s", OpenAPIBaseURL, endpoint, params.Encode())

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return err
	}

	// 添加请求头，禁用压缩
	req.Header.Set("Accept-Encoding", "identity")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 限制响应体大小为 100MB，防止恶意服务器发送大量数据
	limitedReader := io.LimitReader(resp.Body, 100*1024*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return err
	}

	// 先尝试解析为错误
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.ErrCode != 0 {
		if apiErr.ErrCode == 110 {
			return ErrTokenExpired
		}
		return &apiErr
	}

	// 解析结果
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return err
		}
	}

	return nil
}
