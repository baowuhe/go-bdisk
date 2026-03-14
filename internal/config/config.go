package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config CLI配置
type Config struct {
	AppKey    string `mapstructure:"app_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// Manager 配置管理器
type Manager struct {
	configDir string
	v         *viper.Viper
}

// NewManager 创建新的配置管理器
func NewManager() (*Manager, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	m := &Manager{
		configDir: configDir,
		v:         v,
	}

	// 尝试读取配置文件
	_ = m.v.ReadInConfig()

	return m, nil
}

// ConfigDir 获取配置目录
func (m *Manager) ConfigDir() string {
	return m.configDir
}

// TokenPath 获取token文件路径
func (m *Manager) TokenPath() string {
	return filepath.Join(m.configDir, "token.json")
}

// Save 保存配置
func (m *Manager) Save(cfg *Config) error {
	m.v.Set("app_key", cfg.AppKey)
	m.v.Set("secret_key", cfg.SecretKey)
	return m.v.WriteConfigAs(filepath.Join(m.configDir, "config.yaml"))
}

// Load 加载配置
func (m *Manager) Load() (*Config, error) {
	var cfg Config
	err := m.v.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// getConfigDir 获取跨平台的配置目录
func getConfigDir() (string, error) {
	// 首先检查环境变量
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "go-bdisk"), nil
	}

	// 根据操作系统获取标准配置目录
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var configDir string
	switch {
	case os.Getenv("WINDIR") != "": // Windows
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Local")
		}
		configDir = filepath.Join(appData, "go-bdisk")
	case os.Getenv("HOME") != "": // macOS/Linux
		// 检查是否是macOS
		if _, err := os.Stat(filepath.Join(home, "Library", "Application Support")); err == nil {
			configDir = filepath.Join(home, "Library", "Application Support", "go-bdisk")
		} else {
			// Linux
			configDir = filepath.Join(home, ".config", "go-bdisk")
		}
	default:
		// 回退到用户目录下的.go-bdisk
		configDir = filepath.Join(home, ".go-bdisk")
	}

	return configDir, nil
}
