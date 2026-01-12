package decoder

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
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

// Sha1 SHA1 哈希
func Sha1(input string) string {
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

// Sha256 SHA256 哈希
func Sha256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
