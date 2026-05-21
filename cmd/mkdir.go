package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir <path>",
	Short: "创建文件夹",
	Long:  `在百度网盘创建文件夹`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMkdir,
}

func init() {
	mkdirCmd.Flags().IntP("rtype", "r", 1, "文件命名策略：0-不重命名返回冲突，1-冲突即重命名")
	rootCmd.AddCommand(mkdirCmd)
}

func runMkdir(cmd *cobra.Command, args []string) error {
	path := args[0]
	rtype, _ := cmd.Flags().GetInt("rtype")

	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	// 创建文件夹
	if !outputJSON {
		fmt.Printf("正在创建文件夹：%s\n", path)
	}

	result, err := client.File.CreateDir(path, rtype)
	if err != nil {
		if bdisk.IsTokenExpiredError(err) {
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil
		}
		cliutil.PrintError(outputJSON, fmt.Errorf("创建文件夹失败：%w", err))
		return nil
	}

	type resultData struct {
		Path    string `json:"path"`
		FSID    uint64 `json:"fs_id"`
		Message string `json:"message"`
	}

	data := resultData{
		Path:    result.Path,
		FSID:    result.FSID,
		Message: "创建成功",
	}

	cliutil.PrintOutput(outputJSON, data, func() {
		fmt.Printf("创建文件夹成功：%s (fs_id: %d)\n", path, result.FSID)
	})

	return nil
}
