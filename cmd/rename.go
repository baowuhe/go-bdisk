package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var renameCmd = &cobra.Command{
	Use:   "rename <path> <new-name>",
	Short: "重命名文件",
	Long:  `重命名文件或文件夹`,
	Args:  cobra.ExactArgs(2),
	RunE:  runRename,
}

func init() {
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	path := args[0]
	newName := args[1]

	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	// 重命名文件
	if !outputJSON {
		fmt.Printf("正在重命名：%s -> %s\n", path, filepath.Join(filepath.Dir(path), newName))
	}

	err = client.File.Rename(path, newName)
	if err != nil {
		if bdisk.IsTokenExpiredError(err) {
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil
		}
		cliutil.PrintError(outputJSON, fmt.Errorf("重命名失败：%w", err))
		return nil
	}

	type resultData struct {
		Path    string `json:"path"`
		NewName string `json:"new_name"`
		Message string `json:"message"`
	}

	data := resultData{
		Path:    path,
		NewName: newName,
		Message: "重命名成功",
	}

	cliutil.PrintOutput(outputJSON, data, func() {
		fmt.Printf("重命名成功：%s -> %s\n", path, filepath.Join(filepath.Dir(path), newName))
	})

	return nil
}
