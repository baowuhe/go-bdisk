// Package bdisk 提供百度网盘API的Go语言SDK
//
// 该SDK支持设备码授权、文件管理、下载上传等功能。
//
// 快速开始：
//
//	package main
//
//	import (
//		"fmt"
//		"github.com/baowuhe/go-bdisk/pkg/bdisk"
//	)
//
//	func main() {
//		// 创建配置
//		config := bdisk.NewConfig("your_app_key", "your_secret_key")
//		
//		// 创建客户端
//		client, err := bdisk.NewClient(config)
//		if err != nil {
//			panic(err)
//		}
//		
//		// 设备码授权流程
//		deviceResp, err := client.Auth.DeviceCodeFlow()
//		if err != nil {
//			panic(err)
//		}
//		
//		fmt.Printf("请访问: %s\n", deviceResp.VerificationURL)
//		fmt.Printf("验证码: %s\n", deviceResp.UserCode)
//		
//		// 轮询获取token
//		token, err := client.Auth.PollToken(deviceResp.DeviceCode, deviceResp.Interval)
//		if err != nil {
//			panic(err)
//		}
//		
//		client.SetToken(token)
//		fmt.Println("登录成功！")
//	}
//
// Token过期处理：
//
// SDK提供了token过期检测和处理机制：
//
//	// 检查token是否有效
//	if !client.Auth.IsTokenValid() {
//		// token已过期，需要重新登录
//		client.Auth.ClearToken()
//		// 引导用户重新授权...
//	}
//
//	// API调用时也需要处理token过期错误
//	userInfo, err := client.User.GetInfo()
//	if err != nil {
//		if bdisk.IsTokenExpiredError(err) {
//			// token已过期，清除并重新登录
//			client.Auth.ClearToken()
//		}
//	}
//
// 更多示例和文档请访问：https://pkg.go.dev/github.com/baowuhe/go-bdisk/pkg/bdisk
package bdisk
