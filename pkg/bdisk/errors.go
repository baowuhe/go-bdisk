package bdisk

import "errors"

var (
	// ErrTokenExpired 表示访问令牌已过期
	ErrTokenExpired = errors.New("access token has expired")
	
	// ErrNoToken 表示没有找到访问令牌
	ErrNoToken = errors.New("no access token found")
	
	// ErrInvalidConfig 表示配置无效
	ErrInvalidConfig = errors.New("invalid configuration")
)

// APIError 表示百度网盘API返回的错误
type APIError struct {
	ErrCode int    `json:"errno"`
	ErrMsg  string `json:"errmsg"`
	// OAuth标准错误格式
	OAuthErr string `json:"error"`
	ErrDesc  string `json:"error_description"`
}

func (e *APIError) Error() string {
	if e.ErrDesc != "" {
		return e.ErrDesc
	}
	return e.ErrMsg
}

// IsTokenExpiredError 检查错误是否为token过期错误
func IsTokenExpiredError(err error) bool {
	if errors.Is(err, ErrTokenExpired) {
		return true
	}
	
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// 百度网盘API错误码110表示token失效
		return apiErr.ErrCode == 110
	}
	
	return false
}
