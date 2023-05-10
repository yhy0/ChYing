package decoder

import (
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
)

/**
   @author yhy
   @since 2023/5/10
   @desc //TODO
**/

// DecodeUnicode Unicode解码
func DecodeUnicode(str string) string {
	var result strings.Builder
	runes := strings.Split(str, "\\u")
	for _, r := range runes {
		if r == "" {
			continue
		}
		code, err := strconv.ParseInt(r, 16, 32)
		if err != nil {
			result.WriteString(r)
		} else {
			result.WriteRune(rune(code))
		}
	}
	return result.String()
}

// DecodeURL url 解码
func DecodeURL(str string) string {
	decoded, err := url.QueryUnescape(str)
	if err != nil {
		return str
	}
	return decoded
}

// DecodeBase64 Base64 解码
func DecodeBase64(str string) string {
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return str
	}
	return string(decoded)
}

// DecodeHex hex 解码
func DecodeHex(str string) string {
	buf := make([]byte, hex.DecodedLen(len(str)))
	n, err := hex.Decode(buf, []byte(str))

	if err != nil {
		return str
	}
	if n > 0 {
		return string(buf)
	}
	return str
}
