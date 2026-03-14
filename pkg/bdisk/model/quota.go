package model

// QuotaInfo 网盘配额信息
type QuotaInfo struct {
	// Total 总空间大小（字节）
	Total int64 `json:"total"`
	// Used 已用空间大小（字节）
	Used int64 `json:"used"`
	// Free 剩余空间大小（字节）
	Free int64 `json:"free"`
}
