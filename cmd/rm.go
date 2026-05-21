package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var rmCmd = &cobra.Command{
	Use:   "rm <path> [path...]",
	Short: "删除文件",
	Long:  `删除文件或文件夹，支持批量删除`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRm,
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

func runRm(cmd *cobra.Command, args []string) error {
	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	// 删除文件
	paths := args
	if !outputJSON {
		fmt.Printf("正在删除：%s\n", strings.Join(paths, ", "))
	}

	err = client.File.Delete(paths...)
	if err != nil {
		if bdisk.IsTokenExpiredError(err) {
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil
		}
		cliutil.PrintError(outputJSON, fmt.Errorf("删除失败：%w", err))
		return nil
	}

	type resultData struct {
		Paths   []string `json:"paths"`
		Message string   `json:"message"`
	}

	data := resultData{
		Paths:   paths,
		Message: "删除成功",
	}

	cliutil.PrintOutput(outputJSON, data, func() {
		fmt.Printf("删除成功：%s\n", strings.Join(paths, ", "))
	})

	return nil
}
