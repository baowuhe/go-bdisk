package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputJSON bool
	rootCmd    = &cobra.Command{
		Use:   "go-bdisk",
		Short: "百度网盘CLI工具",
		Long:  `使用Go语言开发的百度网盘CLI工具，支持设备码授权、文件管理等功能。`,
	}
)

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&outputJSON, "json", "j", false, "使用JSON格式输出")
}
