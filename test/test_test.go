package test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

/**
  @author: yhy
  @since: 2023/5/7
  @desc: //TODO
**/

func Test(t *testing.T) {
	input := "这是§一个§测试§字符串§ asdasdas "
	re := regexp.MustCompile(`§(.*?)§`)            // 定义正则表达式
	matches := re.FindAllStringSubmatch(input, -1) // 查找所有匹配项

	for _, match := range matches {
		fmt.Println(match[1]) // 输出匹配到的字符串
	}

	str := "foo\rbar\nbaz\r\nqux"

	// 定义分隔符的回调函数
	splitFunc := func(r rune) bool {
		return r == '\r' || r == '\n'
	}

	// 使用 FieldsFunc() 函数分割字符串
	result := strings.FieldsFunc(str, splitFunc)

	// 输出结果字符串切片
	fmt.Println(result)
}
