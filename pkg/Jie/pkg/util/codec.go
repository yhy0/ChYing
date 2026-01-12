package util

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"strings"
)

/**
   @author yhy
   @since 2023/8/20
   @desc //TODO
**/

func interfaceToBytes(i interface{}) []byte {
	var bytes []byte

	switch ret := i.(type) {
	case string:
		bytes = []byte(ret)
	case []byte:
		bytes = ret
	case io.Reader:
		bytes, _ = io.ReadAll(ret)
	default:
		bytes = []byte(fmt.Sprint(i))
	}

	return bytes
}

func DecodeHex(i string) ([]byte, error) {
	return hex.DecodeString(i)
}

func EncodeBase64(i interface{}) string {
	return base64.StdEncoding.EncodeToString(interfaceToBytes(i))
}

func DecodeBase64(i string) ([]byte, error) {
	i = strings.TrimSpace(i)
	i = strings.ReplaceAll(i, "%3d", "=")
	i = strings.ReplaceAll(i, "%3D", "=")

	padding := 4 - len(i)%4
	if padding <= 0 || padding == 4 {
		return base64.StdEncoding.DecodeString(i)
	}
	return base64.StdEncoding.DecodeString(i + strings.Repeat("=", padding))
}
