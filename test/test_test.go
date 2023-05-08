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

func backtrack(cur []string, index int, keys []string, values [][]string, orderedKeys []string) [][]string {
	var results [][]string // 保存结果集合
	if index == len(keys) {
		return [][]string{append([]string{}, cur...)} // 返回当前排列组合
	}
	for _, s := range values[index] {
		cur = append(cur, s)
		subResults := backtrack(cur, index+1, keys, values, orderedKeys)
		results = append(results, subResults...)
		cur = cur[:len(cur)-1]
	}
	return results // 返回所有排列组合
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

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

	m := map[string][]string{
		"A": {"7", "2"},
		"B": {"3", "4", "1"},
	}

	keys := make([]string, 0, len(m))
	values := make([][]string, 0, len(m))
	orderedKeys := make([]string, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
		orderedKeys = append(orderedKeys, k)
	}
	results := backtrack(make([]string, 0, len(keys)), 0, keys, values, orderedKeys)
	for i, result := range results {
		reorderedResult := make([]string, len(result))
		for j, k := range orderedKeys {
			for _, s := range result {
				if m[k] != nil && contains(m[k], s) {
					reorderedResult[j] = s
					break
				}
			}
		}
		results[i] = reorderedResult
	}
	fmt.Printf("%v\n", results) // 返回结果集合
}
