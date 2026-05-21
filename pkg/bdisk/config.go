package bdisk

// Config 百度网盘SDK配置
type Config struct {
	// AppKey 应用Key
	AppKey string
	
	// SecretKey 应用密钥
	SecretKey string
}

// NewConfig 创建新的SDK配置
func NewConfig(appKey, secretKey string) *Config {
	return &Config{
		AppKey:    appKey,
		SecretKey: secretKey,
	}
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	if c.AppKey == "" {
		return ErrInvalidConfig
	}
	if c.SecretKey == "" {
		return ErrInvalidConfig
	}
	return nil
}
