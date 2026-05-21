package cmd

import (
	"fmt"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/internal/config"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

// initClient 初始化 SDK 客户端，处理登录检查和 token 刷新
// 返回 client、cfgManager 和 cfg，如果未登录或初始化失败则返回错误并打印提示
func initClient() (*bdisk.Client, *config.Manager, *config.Config, error) {
	// 初始化配置管理器
	cfgManager, err := config.NewManager()
	if err != nil {
		cliutil.PrintError(outputJSON, fmt.Errorf("初始化配置失败：%w", err))
		return nil, nil, nil, fmt.Errorf("初始化配置失败：%w", err)
	}

	// 加载配置
	cfg, err := cfgManager.Load()
	if err != nil {
		cliutil.PrintError(outputJSON, fmt.Errorf("加载配置失败：%w", err))
		return nil, nil, nil, fmt.Errorf("加载配置失败：%w", err)
	}

	if cfg.AppKey == "" || cfg.SecretKey == "" {
		cliutil.PrintError(outputJSON, fmt.Errorf("请先配置应用 Key 和密钥"))
		return nil, nil, nil, fmt.Errorf("请先配置应用 Key 和密钥：go-bdisk login --app-key <key> --secret-key <secret>")
	}

	// 加载 token
	token, err := cfgManager.LoadToken()
	if err != nil {
		cliutil.PrintError(outputJSON, fmt.Errorf("请先登录：%w", err))
		return nil, nil, nil, fmt.Errorf("请先登录：go-bdisk login")
	}

	// 创建 SDK 客户端
	bdConfig := bdisk.NewConfig(cfg.AppKey, cfg.SecretKey)
	client, err := bdisk.NewClient(bdConfig)
	if err != nil {
		cliutil.PrintError(outputJSON, fmt.Errorf("创建客户端失败：%w", err))
		return nil, nil, nil, fmt.Errorf("创建客户端失败：%w", err)
	}
	client.SetToken(token)

	// 检查 token 有效性，如果已过期则尝试刷新
	if !client.Auth.IsTokenValid() {
		if outputJSON {
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，尝试刷新令牌..."))
		} else {
			fmt.Println("登录已过期，尝试刷新令牌...")
		}

		// 尝试使用 refresh_token 刷新
		newToken, err := client.Auth.RefreshToken(token.RefreshToken)
		if err != nil {
			// 刷新失败，清除 token
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil, nil, nil, fmt.Errorf("登录已过期，请重新登录")
		}

		// 保存新 token
		if err := cfgManager.SaveToken(newToken); err != nil {
			cliutil.PrintError(outputJSON, fmt.Errorf("保存新令牌失败：%w", err))
			return nil, nil, nil, fmt.Errorf("保存新令牌失败：%w", err)
		}
		client.SetToken(newToken)

		if !outputJSON {
			fmt.Println("令牌刷新成功！")
		}
	}

	return client, cfgManager, cfg, nil
}
