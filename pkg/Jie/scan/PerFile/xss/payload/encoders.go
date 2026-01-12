package payload

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
)

// EncodingType 定义了编码类型的枚举
type EncodingType int

const (
	HTMLEncoding EncodingType = iota
	URLEncoding
	UnicodeEncoding
	HexEncoding
	Base64Encoding
)

// Encode 统一的编码入口
func Encode(input string, encodingType EncodingType) string {
	switch encodingType {
	case HTMLEncoding:
		return htmlEncode(input)
	case URLEncoding:
		return urlEncode(input, false)
	case UnicodeEncoding:
		return unicodeEncode(input)
	case HexEncoding:
		return hexEncode(input)
	default:
		return input
	}
}

// htmlEncode 对字符串进行HTML实体编码（十进制和十六进制混合）
func htmlEncode(input string) string {
	var buffer bytes.Buffer
	for i, r := range input {
		// 混合使用十进制和十六进制实体
		if i%2 == 0 {
			buffer.WriteString(fmt.Sprintf("&#%d;", r))
		} else {
			buffer.WriteString(fmt.Sprintf("&#x%x;", r))
		}
	}
	return buffer.String()
}

// urlEncode 对字符串进行URL编码
// full=true 会编码所有字符，否则只编码特殊字符
func urlEncode(input string, full bool) string {
	var result string
	if full {
		for _, b := range []byte(input) {
			result += fmt.Sprintf("%%%02x", b)
		}
	} else {
		result = url.QueryEscape(input)
	}
	return result
}

// unicodeEncode 将字符串编码为JS Unicode转义序列
func unicodeEncode(input string) string {
	var buffer bytes.Buffer
	for _, r := range input {
		buffer.WriteString(fmt.Sprintf("\\u%04x", r))
	}
	return buffer.String()
}

// hexEncode 将字符串编码为JS Hex转义序列
func hexEncode(input string) string {
	var buffer bytes.Buffer
	for _, r := range input {
		buffer.WriteString(fmt.Sprintf("\\x%02x", r))
	}
	return buffer.String()
}

// MixedEncode 混合编码，用于绕过复杂的过滤器
func MixedEncode(payload string) []string {
	var variants []string
	// 示例：部分URL编码，部分HTML编码
	if len(payload) > 4 {
		mid := len(payload) / 2
		part1 := payload[:mid]
		part2 := payload[mid:]
		variants = append(variants, urlEncode(part1, false)+htmlEncode(part2))
		variants = append(variants, htmlEncode(part1)+urlEncode(part2, false))
	}

	// 示例：随机字符编码
	var sb strings.Builder
	for i, r := range payload {
		switch i % 4 {
		case 0:
			sb.WriteRune(r)
		case 1:
			sb.WriteString(urlEncode(string(r), true))
		case 2:
			sb.WriteString(htmlEncode(string(r)))
		case 3:
			sb.WriteString(unicodeEncode(string(r)))
		}
	}
	variants = append(variants, sb.String())

	return variants
}
