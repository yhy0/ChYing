package decoder

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"net/url"
	"strings"
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
	encoded := url.QueryEscape(str)                   // 这个函数会将空格，替换为+
	encoded = strings.ReplaceAll(encoded, "+", "%20") // 这里再将+ 替换为 %2B
	return encoded
}

// EncodeBase64 Base64 编码
func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// EncodeHex hex 编码
func EncodeHex(str string) string {
	buf := make([]byte, hex.EncodedLen(len(str)))
	hex.Encode(buf, []byte(str))
	return string(buf)
}

// EncodeHTML HTML 编码
func EncodeHTML(str string) string {
	return html.EscapeString(str)
}
