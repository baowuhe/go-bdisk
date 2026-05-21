package bdisk

import (
	"net/url"

	"github.com/baowuhe/go-bdisk/pkg/bdisk/model"
)

// GetInfo 获取用户信息
func (u *UserService) GetInfo() (*model.UserInfo, error) {
	var result struct {
		BaiduName   string `json:"baidu_name"`
		NetdiskName string `json:"netdisk_name"`
		AvatarURL   string `json:"avatar_url"`
		Vip         int    `json:"vip"`
	}

	err := u.client.doRequest("GET", "/xpan/nas", url.Values{
		"method": {"uinfo"},
	}, &result)

	if err != nil {
		return nil, err
	}

	return &model.UserInfo{
		BaiduName:   result.BaiduName,
		NetdiskName: result.NetdiskName,
		AvatarURL:   result.AvatarURL,
		Vip:         result.Vip,
	}, nil
}

// GetQuota 获取配额信息
func (u *UserService) GetQuota() (*model.QuotaInfo, error) {
	var result struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
		Free  int64 `json:"free"`
	}

	err := u.client.doRequest("GET", "/xpan/nas", url.Values{
		"method": {"quota"},
	}, &result)

	if err != nil {
		return nil, err
	}

	return &model.QuotaInfo{
		Total: result.Total,
		Used:  result.Used,
		Free:  result.Free,
	}, nil
}
