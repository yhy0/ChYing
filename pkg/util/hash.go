package util

import (
	"crypto/md5"
	"encoding/hex"
)

/**
  @author: yhy
  @since: 2023/5/7
  @desc: //TODO
**/

// Md5 加密
func Md5(input string) string {
	// 计算 MD5 哈希值
	hash := md5.Sum([]byte(input))

	// 将哈希值转换为十六进制字符串
	hexHash := hex.EncodeToString(hash[:])

	return hexHash
}
