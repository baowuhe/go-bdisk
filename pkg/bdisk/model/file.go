package model

import "time"

// FileType 文件类型
type FileType string

const (
	// FileTypeFile 文件
	FileTypeFile FileType = "file"
	// FileTypeDir 文件夹
	FileTypeDir FileType = "dir"
)

// FileInfo 文件信息
type FileInfo struct {
	// FSID 文件唯一标识
	FSID int64 `json:"fs_id"`
	// Path 文件路径
	Path string `json:"path"`
	// Name 文件名
	Name string `json:"name"`
	// Type 文件类型
	Type FileType `json:"type"`
	// Size 文件大小（字节）
	Size int64 `json:"size"`
	// ModifiedAt 修改时间
	ModifiedAt time.Time `json:"modified_at"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	// MD5 文件MD5
	MD5 string `json:"md5"`
}

// FileList 文件列表响应
type FileList struct {
	// Path 当前路径
	Path string `json:"path"`
	// Items 文件列表
	Items []*FileInfo `json:"items"`
	// Total 总数
	Total int `json:"total"`
}
