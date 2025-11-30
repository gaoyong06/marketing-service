package biz

import "github.com/google/uuid"

// GenerateShortID 生成短 ID（32个字符，去掉 UUID 的连字符）
// 用于生成符合数据库 varchar(32) 字段要求的 ID
func GenerateShortID() string {
	id := uuid.New().String()
	// UUID 格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36个字符)
	// 去掉连字符后: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx (32个字符)
	return id[:8] + id[9:13] + id[14:18] + id[19:23] + id[24:36]
}
