package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baowuhe/go-bdisk/internal/config"
	"github.com/baowuhe/go-bdisk/pkg/bdisk"
)

var (
	appKey    string
	secretKey string

	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "登录百度网盘",
		Long:  `使用设备码授权方式登录百度网盘`,
		RunE:  runLogin,
	}
)

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVar(&appKey, "app-key", "", "百度应用Key")
	loginCmd.Flags().StringVar(&secretKey, "secret-key", "", "百度应用密钥")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// 初始化配置管理器
	cfgManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}

	// 加载或设置配置
	var cfg *config.Config
	if appKey != "" && secretKey != "" {
		cfg = &config.Config{
			AppKey:    appKey,
			SecretKey: secretKey,
		}
		if err := cfgManager.Save(cfg); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
	} else {
		cfg, err = cfgManager.Load()
		if err != nil || cfg.AppKey == "" || cfg.SecretKey == "" {
			return fmt.Errorf("请先配置应用Key和密钥: go-bdisk login --app-key <key> --secret-key <secret>")
		}
	}

	// 创建SDK客户端
	bdConfig := bdisk.NewConfig(cfg.AppKey, cfg.SecretKey)
	client, err := bdisk.NewClient(bdConfig)
	if err != nil {
		return fmt.Errorf("创建客户端失败: %w", err)
	}

	// 获取设备码
	deviceResp, err := client.Auth.DeviceCodeFlow()
	if err != nil {
		return fmt.Errorf("获取设备码失败: %w", err)
	}

	// 显示验证信息
	if outputJSON {
		fmt.Printf(`{"success":true,"data":{"verification_url":"%s","user_code":"%s"}}`+"\n",
			deviceResp.VerificationURL, deviceResp.UserCode)
	} else {
		fmt.Println("=")
		fmt.Println("请使用浏览器访问以下链接并输入验证码：")
		fmt.Printf("验证链接: %s\n", deviceResp.VerificationURL)
		fmt.Printf("验证码: %s\n", deviceResp.UserCode)
		fmt.Println("=")
		fmt.Println("正在等待授权...")
	}

	// 轮询获取token
	token, err := client.Auth.PollToken(deviceResp.DeviceCode, deviceResp.Interval)
	if err != nil {
		return fmt.Errorf("获取令牌失败: %w", err)
	}

	// 保存token
	if err := cfgManager.SaveToken(token); err != nil {
		return fmt.Errorf("保存令牌失败: %w", err)
	}

	if outputJSON {
		fmt.Println(`{"success":true,"data":{"message":"登录成功"}}`)
	} else {
		fmt.Println("登录成功！")
	}

	return nil
}
