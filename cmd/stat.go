package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var statCmd = &cobra.Command{
	Use:   "stat <path>",
	Short: "查看文件信息",
	Long:  `查看指定文件或文件夹的详细信息`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStat,
}

func init() {
	rootCmd.AddCommand(statCmd)
}

func runStat(cmd *cobra.Command, args []string) error {
	path := args[0]

	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	// 获取文件信息
	fileInfo, err := client.File.GetInfo(path)
	if err != nil {
		if bdisk.IsTokenExpiredError(err) {
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil
		}
		cliutil.PrintError(outputJSON, fmt.Errorf("获取文件信息失败：%w", err))
		return nil
	}

	cliutil.PrintOutput(outputJSON, fileInfo, func() {
		fmt.Println("文件信息:")
		fmt.Printf("  路径：%s\n", fileInfo.Path)
		fmt.Printf("  名称：%s\n", fileInfo.Name)
		fmt.Printf("  类型：%s\n", fileInfo.Type)
		if fileInfo.Type != "dir" {
			fmt.Printf("  大小：%s\n", cliutil.FormatSize(fileInfo.Size))
		}
		fmt.Printf("  FSID：%d\n", fileInfo.FSID)
		fmt.Printf("  创建时间：%s\n", fileInfo.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  修改时间：%s\n", fileInfo.ModifiedAt.Format("2006-01-02 15:04:05"))
		if fileInfo.MD5 != "" {
			fmt.Printf("  MD5: %s\n", fileInfo.MD5)
		}
	})

	return nil
}
