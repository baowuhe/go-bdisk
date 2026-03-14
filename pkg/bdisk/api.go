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

	body, err := io.ReadAll(resp.Body)
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
