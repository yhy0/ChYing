package test

import (
	"fmt"
	"github.com/yhy0/ChYing/tools/decoder"
	"testing"
)

/**
   @author yhy
   @since 2023/5/10
   @desc //TODO
**/

func TestDecode(t *testing.T) {
	fmt.Println(decoder.DecodeUnicode("\\u0048\\u0065\\u006c\\u006c\\u006f\\u002c\\u0020\\u4e16\\u754c\\u0021"))
	fmt.Println(decoder.DecodeURL("Hello%2C%20%E4%B8%96%E7%95%8C%2B%21"))
	fmt.Println(decoder.DecodeBase64("SGVsbG8sIOS4lueVjCE="))
	fmt.Println(decoder.DecodeHex("48656c6c6f2c20e4b896e7958c21"))
	fmt.Println(decoder.DecodeUnicode(`@\u006fgnl.OgnlC\u006fntext`))
}

func TestEncode(t *testing.T) {
	fmt.Println(decoder.EncodeUnicode("Hello, 世界!"))
	fmt.Println(decoder.EncodeURL("Hello, 世界+!"))
	fmt.Println(decoder.EncodeBase64("Hello, 世界!"))
	fmt.Println(decoder.EncodeHex("Hello, 世界!"))
}
