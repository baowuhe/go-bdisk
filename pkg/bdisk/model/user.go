package model

// UserInfo 用户信息
type UserInfo struct {
	// BaiduName 百度账号
	BaiduName string `json:"baidu_name"`
	// NetdiskName 网盘账号
	NetdiskName string `json:"netdisk_name"`
	// AvatarURL 头像地址
	AvatarURL string `json:"avatar_url"`
	// VIP 会员类型
	Vip int `json:"vip"`
}
