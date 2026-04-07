package bdisk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"
)

const (
	// OAuthBaseURL 百度 OAuth 基础 URL
	OAuthBaseURL = "https://openapi.baidu.com/oauth/2.0"
	// OpenAPIBaseURL 百度网盘 OpenAPI 基础 URL
	OpenAPIBaseURL = "https://pan.baidu.com/rest/2.0"
)

// Token 访问令牌
type Token struct {
	// AccessToken 访问令牌
	AccessToken string `json:"access_token"`
	// RefreshToken 刷新令牌
	RefreshToken string `json:"refresh_token"`
	// ExpiresIn 有效期（秒）
	ExpiresIn int64 `json:"expires_in"`
	// CreatedAt 创建时间戳（Unix 时间戳）
	CreatedAt int64 `json:"created_at"`
	// ExpiresAt 过期时间（运行时计算，不保存到 JSON）
	ExpiresAt time.Time `json:"-"`
}

// IsExpired 检查 token 是否已过期
func (t *Token) IsExpired() bool {
	if t.CreatedAt == 0 {
		return true
	}
	expiresAt := time.Unix(t.CreatedAt, 0).Add(time.Duration(t.ExpiresIn) * time.Second)
	return time.Now().After(expiresAt)
}

// IsValid 检查 token 是否有效
func (t *Token) IsValid() bool {
	if t == nil || t.AccessToken == "" {
		return false
	}
	return !t.IsExpired()
}

// DeviceCodeResponse 设备码响应
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	QrcodeURL       string `json:"qrcode_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// DeviceCodeFlow 设备码授权流程
func (a *AuthService) DeviceCodeFlow() (*DeviceCodeResponse, error) {
	params := url.Values{}
	params.Set("response_type", "device_code")
	params.Set("client_id", a.client.config.AppKey)
	params.Set("scope", "basic,netdisk")

	reqURL := fmt.Sprintf("%s/device/code?%s", OAuthBaseURL, params.Encode())

	resp, err := a.client.http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		DeviceCodeResponse
		*APIError
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.APIError != nil && result.ErrCode != 0 {
		return nil, result.APIError
	}

	return &result.DeviceCodeResponse, nil
}

// PollToken 轮询获取令牌，直到成功获取token
func (a *AuthService) PollToken(deviceCode string, interval int) (*Token, error) {
	// 轮询间隔5秒（百度API要求最小5秒）
	pollInterval := 5
	if pollInterval < 5 {
		pollInterval = 5
	}

	pollCount := 0
	for {
		pollCount++
		// 倒计时显示
		for i := pollInterval; i > 0; i-- {
			fmt.Printf("\r等待授权中 [第%d次查询结果，%d秒后进行下一次] ...", pollCount, i)
			time.Sleep(1 * time.Second)
		}

		token, _ := a.getToken(deviceCode)

		// 成功获取 token
		if token != nil && token.AccessToken != "" {
			fmt.Printf("\r等待授权中 [第%d次查询结果，授权成功]          \n", pollCount)
			return token, nil
		}
	}
}

func (a *AuthService) getToken(deviceCode string) (*Token, error) {
	params := url.Values{}
	params.Set("grant_type", "device_token")
	params.Set("code", deviceCode)
	params.Set("client_id", a.client.config.AppKey)
	params.Set("client_secret", a.client.config.SecretKey)

	reqURL := fmt.Sprintf("%s/token?%s", OAuthBaseURL, params.Encode())

	resp, err := a.client.http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 先尝试解析为错误（支持百度errno格式和OAuth标准error格式）
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil {
		// 检查百度自定义errno格式
		if apiErr.ErrCode != 0 {
			return nil, &apiErr
		}
		// 检查OAuth标准error格式
		if apiErr.OAuthErr != "" {
			// slow_down不是失败，只是让减慢轮询速度，返回nil让调用方继续轮询
			if apiErr.OAuthErr == "slow_down" {
				return nil, nil
			}
			return nil, &apiErr
		}
	}

	// 解析 token 响应
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		// JSON解析失败，继续轮询
		return nil, nil
	}

	// 如果没有获取到 token，返回 nil 而不是错误，让调用者继续轮询
	if tokenResp.AccessToken == "" {
		return nil, nil
	}

	token := &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		CreatedAt:    time.Now().Unix(),
	}

	return token, nil
}

// IsTokenValid 检查当前 token 是否有效
func (a *AuthService) IsTokenValid() bool {
	return a.client.token.IsValid()
}

// ClearToken 清除 token
func (a *AuthService) ClearToken() {
	a.client.ClearToken()
}

// RefreshToken 使用 refresh_token 刷新 access_token
func (a *AuthService) RefreshToken(refreshToken string) (*Token, error) {
	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("refresh_token", refreshToken)
	params.Set("client_id", a.client.config.AppKey)
	params.Set("client_secret", a.client.config.SecretKey)

	reqURL := fmt.Sprintf("%s/token?%s", OAuthBaseURL, params.Encode())

	resp, err := a.client.http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 先尝试解析为错误
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.ErrCode != 0 {
		return nil, &apiErr
	}

	// 解析 token 响应
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("refresh token failed: empty access token")
	}

	token := &Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		CreatedAt:    time.Now().Unix(),
	}

	return token, nil
}
