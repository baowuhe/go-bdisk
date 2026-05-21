package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "列出文件",
	Long:  `列出指定路径下的文件和文件夹，默认为根目录`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLs,
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

func runLs(cmd *cobra.Command, args []string) error {
	path := "/"
	if len(args) > 0 {
		path = args[0]
	}

	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	// 获取文件列表
	fileList, err := client.File.List(path)
	if err != nil {
		if bdisk.IsTokenExpiredError(err) {
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil
		}
		cliutil.PrintError(outputJSON, fmt.Errorf("获取文件列表失败：%w", err))
		return nil
	}

	cliutil.PrintOutput(outputJSON, fileList, func() {
		fmt.Printf("路径：%s\n", fileList.Path)
		fmt.Printf("文件总数：%d\n", fileList.Total)
		fmt.Println()

		for _, item := range fileList.Items {
			typeLabel := "[文件]"
			if item.Type == "dir" {
				typeLabel = "[文件夹]"
			}

			sizeStr := ""
			if item.Type != "dir" {
				sizeStr = cliutil.FormatSize(item.Size)
			}

			timeStr := item.ModifiedAt.Format("2006-01-02 15:04")
			fmt.Printf("%s  %-20s %-10s %s\n", typeLabel, item.Name, sizeStr, timeStr)
		}
	})

	return nil
}
