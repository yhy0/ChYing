package test

import (
	"encoding/json"
	"fmt"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/ChYing/tools"
	"regexp"
	"strings"
	"testing"
)

/*
*

	@author: yhy
	@since: 2023/5/7
	@desc: //TODO

*
*/
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

	arr := [][]string{
		{"a", "b"},
		{"x", "y", "z"},
		{"alpha", "beta"},
	}

	res := make([][]string, len(arr[0]))

	for k := range arr[0] {
		res[k] = []string{arr[0][k]}

		for i := range arr {
			if i > 0 {
				res[k] = append(res[k], arr[i][k])
			}
		}

	}
	fmt.Println(res)

}

func TestAv(t *testing.T) {

	// 读取 av.json 文件内容
	data, err := file.AvFile.ReadFile("av.json")
	if err != nil {
		panic(err)
	}
	// 解析 JSON 数据到一个 map 对象中
	file.Av = make(map[string]string)
	err = json.Unmarshal(data, &file.Av)
	if err != nil {
		panic(err)
	}

	out := `
svchost.exe                   2912 Services                   0      6,620 K
svchost.exe                   2968 Services                   0      1,388 K
svchost.exe                   2976 Services                   0      6,960 K
svchost.exe                   3032 Services                   0      3,032 K
svchost.exe                   3060 Services                   0      1,772 K
Memory Compression            2140 Services                   0    696,916 K
svchost.exe                   3084 Services                   0      4,236 K
svchost.exe                   3128 Services                   0      2,912 K
svchost.exe                   3184 Services                   0      3,468 K
NVDisplay.Container.exe       3392 Console                    1     23,808 K
svchost.exe                   3492 Services                   0      9,716 K
svchost.exe                   3516 Services                   0     17,464 K
svchost.exe                   3636 Services                   0      2,872 K
svchost.exe                   3644 Services                   0      5,816 K
svchost.exe                   3760 Services                   0      4,584 K
svchost.exe                   3860 Services                   0     10,464 K
wlanext.exe                   4024 Services                   0      1,808 K
conhost.exe                   4032 Services                   0      1,012 K
usysdiag.exe                  3684 Services                   0        852 K
ZhuDongFangYu.exe             4164 Services                   0     14,324 K
svchost.exe                   4356 Services                   0      9,020 K
spoolsv.exe                   4460 Services                   0      5,168 K
wsctrlsvc.exe                 4588 Services                   0      4,964 K
svchost.exe                   4736 Services                   0     13,712 K
svchost.exe                   4752 Services                   0      5,084 K
svchost.exe                   5056 Services                   0      3,316 K
svchost.exe         
`
	fmt.Println(tools.Tasklist(out))
}
