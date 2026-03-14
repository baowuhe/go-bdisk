package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var cpCmd = &cobra.Command{
	Use:   "cp <source> <destination>",
	Short: "复制文件",
	Long:  `复制文件或文件夹到指定位置`,
	Args:  cobra.ExactArgs(2),
	RunE:  runCp,
}

func init() {
	cpCmd.Flags().StringP("ondup", "o", "fail", "重名处理策略：fail(失败), newcopy(覆盖), skip(跳过)")
	rootCmd.AddCommand(cpCmd)
}

func runCp(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	destPath := args[1]

	ondup, _ := cmd.Flags().GetString("ondup")

	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	// 复制文件
	if !outputJSON {
		fmt.Printf("正在复制：%s -> %s\n", srcPath, destPath)
	}

	err = client.File.Copy(srcPath, destPath, ondup)
	if err != nil {
		if bdisk.IsTokenExpiredError(err) {
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil
		}
		cliutil.PrintError(outputJSON, fmt.Errorf("复制文件失败：%w", err))
		return nil
	}

	type resultData struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Message     string `json:"message"`
	}

	data := resultData{
		Source:      srcPath,
		Destination: destPath,
		Message:     "复制成功",
	}

	cliutil.PrintOutput(outputJSON, data, func() {
		fmt.Printf("复制成功：%s -> %s\n", srcPath, destPath)
	})

	return nil
}
