package test

import (
	"encoding/json"
	"fmt"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/ChYing/tools/burpSuite"
	"github.com/yhy0/ChYing/tools/gadget"
	"github.com/yhy0/logging"
	"regexp"
	"testing"
	"time"
)

/**
	@author: yhy
	@since: 2023/5/7
	@desc: //TODO
**/

func Test(t *testing.T) {
	pattern := `www.google.com`

	domain1 := "123123.google.com"
	domain2 := "example.google.net"
	domain3 := "www.google.com"
	domain4 := "google.co"

	match1, _ := regexp.MatchString(pattern, domain1)
	match2, _ := regexp.MatchString(pattern, domain2)
	match3, _ := regexp.MatchString(pattern, domain3)
	match4, _ := regexp.MatchString(pattern, domain4)

	fmt.Println(match1)
	fmt.Println(match2)
	fmt.Println(match3)
	fmt.Println(match4)

	logging.New(true, "", "test")
	burpSuite.Init()
	fmt.Println(burpSuite.Settings)
	burpSuite.HotConf()
	time.Sleep(3 * time.Second)

	Settings := &burpSuite.Setting{
		ProxyPort: 119080,
		Exclude: []string{
			`^.*\.google.*$`,
			`^.*\.baidu.*$`,
			`^.*\.doubleclick.*$`,
		},
		Include: []string{
			`baidu.com`,
		},
	}

	exclude := ""

	for _, e := range Settings.Exclude {
		exclude += fmt.Sprintf("  - %s\r\n", e)
	}

	include := ""

	if len(Settings.Include) == 0 {
		include = "  - "
	} else {
		for _, i := range Settings.Include {
			include += fmt.Sprintf("  - %s\r\n", i)
		}

	}

	var defaultYamlByte = []byte(fmt.Sprintf(`
port: %d
exclude:
%s
include:
%s
`, Settings.ProxyPort, exclude, include))

	burpSuite.WriteYamlConfig(defaultYamlByte)

	fmt.Println(burpSuite.Settings)

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
	fmt.Println(gadget.Tasklist(out))

	port, err := utils.GetRandomUnusedPort()

	fmt.Println(port)
}
