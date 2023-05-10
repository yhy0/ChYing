package decoder

import (
	"encoding/base64"
	"fmt"
	"net/url"
)

/**
   @author yhy
   @since 2023/5/10
   @desc //TODO
**/

// EncodeUnicode Unicode编码
func EncodeUnicode(str string) string {
	var result string
	for _, r := range str {
		result += "\\u" + fmt.Sprintf("%04x", r)
	}
	return result
}

// EncodeURL URL 编码
func EncodeURL(str string) string {
	return url.QueryEscape(str)
}

// EncodeBase64 Base64 编码
func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
