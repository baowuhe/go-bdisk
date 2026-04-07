package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <local-path> [remote-path]",
	Short: "上传文件",
	Long:  `上传本地文件到百度网盘`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, args []string) error {
	localPath := args[0]
	remotePath := filepath.Join("/", filepath.Base(localPath)) // 默认上传到根目录

	if len(args) > 1 {
		remotePath = args[1]
	}

	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	if !outputJSON {
		fmt.Printf("正在上传：%s -> %s\n", localPath, remotePath)
		fmt.Println()
	}

	// 上传文件
	actualRemotePath, err := client.Upload.StartWithProgress(localPath, remotePath, func(progress bdisk.UploadProgress) {
		if outputJSON {
			// JSON 格式：打印 JSON 行
			jsonProgress := map[string]interface{}{
				"uploaded":     progress.Uploaded,
				"total":        progress.Total,
				"percent":      fmt.Sprintf("%.2f", progress.Percent),
				"current_part": progress.CurrentPart,
				"total_parts":  progress.TotalParts,
			}
			jsonData, _ := json.Marshal(jsonProgress)
			fmt.Println(string(jsonData))
		} else {
			// 非 JSON 格式：同一行内打印进度
			percentStr := fmt.Sprintf("%.2f", progress.Percent)
			msg := fmt.Sprintf("\r%s / %s [%s%%] 分片：%d/%d",
				formatBytes(progress.Uploaded),
				formatBytes(progress.Total),
				percentStr,
				progress.CurrentPart,
				progress.TotalParts)
			fmt.Print(msg)
		}
	})

	if !outputJSON {
		fmt.Println() // 换行
	}

	if err != nil {
		if bdisk.IsTokenExpiredError(err) {
			client.Auth.ClearToken()
			cfgManager.ClearToken()
			cliutil.PrintError(outputJSON, fmt.Errorf("登录已过期，请重新登录"))
			return nil
		}
		cliutil.PrintError(outputJSON, fmt.Errorf("上传失败：%w", err))
		return nil
	}

	type resultData struct {
		LocalPath  string `json:"local_path"`
		RemotePath string `json:"remote_path"`
		Message    string `json:"message"`
	}

	data := resultData{
		LocalPath:  localPath,
		RemotePath: actualRemotePath,
		Message:    "上传成功",
	}

	cliutil.PrintOutput(outputJSON, data, func() {
		fmt.Printf("上传成功：%s -> %s\n", localPath, actualRemotePath)
	})

	return nil
}
