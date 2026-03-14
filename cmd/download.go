package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/cliutil"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var downloadCmd = &cobra.Command{
	Use:   "download <remote-path> [local-path]",
	Short: "下载文件",
	Long:  `从百度网盘下载文件到本地`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}

func runDownload(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

	// 初始化客户端（自动处理登录检查和 token 刷新）
	client, cfgManager, _, err := initClient()
	if err != nil {
		return nil
	}

	// 获取文件信息以确定文件名
	_, _, fileName, err := client.Download.GetFileFSID(remotePath)
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

	// 确定本地路径
	localPath := fileName // 默认使用网盘文件名
	if len(args) > 1 {
		localPath = args[1]
	}

	if !outputJSON {
		fmt.Printf("正在下载：%s -> %s\n", remotePath, localPath)
		fmt.Println()
	}

	// 下载文件
	var progressMessages []string
	err = client.Download.StartWithProgress(remotePath, localPath, func(progress bdisk.DownloadProgress) {
		if outputJSON {
			// JSON 格式：打印 JSON 行
			jsonProgress := map[string]interface{}{
				"downloaded": progress.Downloaded,
				"total":      progress.Total,
				"percent":    fmt.Sprintf("%.2f", progress.Percent),
			}
			jsonData, _ := json.Marshal(jsonProgress)
			fmt.Println(string(jsonData))
		} else {
			// 非 JSON 格式：同一行内打印进度
			percentStr := fmt.Sprintf("%.2f", progress.Percent)
			msg := fmt.Sprintf("\r%s / %s [%s%%]",
				formatBytes(progress.Downloaded),
				formatBytes(progress.Total),
				percentStr)
			fmt.Print(msg)
			progressMessages = append(progressMessages, msg)
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
		cliutil.PrintError(outputJSON, fmt.Errorf("下载失败：%w", err))
		return nil
	}

	type resultData struct {
		RemotePath string `json:"remote_path"`
		LocalPath  string `json:"local_path"`
		Message    string `json:"message"`
	}

	data := resultData{
		RemotePath: remotePath,
		LocalPath:  localPath,
		Message:    "下载成功",
	}

	cliutil.PrintOutput(outputJSON, data, func() {
		fmt.Printf("下载成功：%s -> %s\n", remotePath, localPath)
	})

	return nil
}

// formatBytes 格式化字节数为人类可读格式
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
