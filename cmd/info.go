package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "查询用户信息和网盘配额",
	Long:  `查询当前登录用户的信息和网盘配额使用情况`,
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	// 获取用户信息
	userInfo, err := client.User.GetInfo()
	if err != nil {
		if bdisk.IsTokenExpiredError(err) {
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil
		}
		cliutil.PrintError(outputJSON, fmt.Errorf("获取用户信息失败：%w", err))
		return nil
	}

	// 准备数据用于输出
	type infoData struct {
		User interface{} `json:"user"`
	}

	data := infoData{
		User: userInfo,
	}

	cliutil.PrintOutput(outputJSON, data, func() {
		fmt.Println("用户信息:")
		fmt.Printf("  百度账号：%s\n", userInfo.BaiduName)
		fmt.Printf("  网盘账号：%s\n", userInfo.NetdiskName)
		fmt.Printf("  会员类型：%d\n", userInfo.Vip)
		fmt.Println()
		fmt.Println("提示：配额信息功能暂不可用")
	})

	return nil
}
