package cliutil

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONResponse 统一JSON响应结构
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PrintOutput 根据outputJSON标志输出内容
func PrintOutput(outputJSON bool, data interface{}, humanReadable func()) {
	if outputJSON {
		resp := JSONResponse{
			Success: true,
			Data:    data,
		}
		json.NewEncoder(os.Stdout).Encode(resp)
	} else {
		humanReadable()
	}
}

// PrintError 输出错误
func PrintError(outputJSON bool, err error) {
	if outputJSON {
		resp := JSONResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(os.Stdout).Encode(resp)
	} else {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
	}
	os.Exit(1)
}

// FormatSize 格式化文件大小
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
