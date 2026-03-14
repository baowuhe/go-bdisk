package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/internal/config"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "退出登录",
	Long:  `退出登录并清空所有缓存在本地的凭证数据`,
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	// 初始化配置管理器
	cfgManager, err := config.NewManager()
	if err != nil {
		cliutil.PrintError(outputJSON, fmt.Errorf("初始化配置失败：%w", err))
		return nil
	}

	// 清除 token
	if err := cfgManager.ClearToken(); err != nil {
		cliutil.PrintError(outputJSON, fmt.Errorf("清除令牌失败：%w", err))
		return nil
	}

	// 如果有客户端，也清除内存中的 token
	cfg, err := cfgManager.Load()
	if err == nil && cfg.AppKey != "" && cfg.SecretKey != "" {
		bdConfig := bdisk.NewConfig(cfg.AppKey, cfg.SecretKey)
		client, err := bdisk.NewClient(bdConfig)
		if err == nil {
			client.Auth.ClearToken()
		}
	}

	cliutil.PrintOutput(outputJSON, map[string]string{"message": "已退出登录"}, func() {
		fmt.Println("已退出登录，所有本地凭证数据已清空")
	})

	return nil
}
