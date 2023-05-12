package tools

import (
	"github.com/yhy0/ChYing/pkg/file"
	"strings"
)

/**
   @author yhy
   @since 2023/5/12
   @desc https://github.com/gh0stkey/avList/blob/master/avlist.js
**/

func Tasklist(out string) map[string]string {
	// 将命令输出结果转换为字符串类型，并按行切分
	lines := strings.Split(out, "\n")
	var res = make(map[string]string)
	// 遍历每一行输出，查找包含在 map 中的进程名称，并输出对应的描述信息
	for _, line := range lines {
		// 将每一行输出按照空格切分，得到进程相关信息
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// 获取进程名称，并在 map 中查找对应的描述信息
		name := fields[0]

		desc, ok := file.Av[name]
		if !ok {
			// 进程名称不在 map 中，忽略该进程
			continue
		}
		// 输出进程名称和描述信息
		res[name] = desc
	}
	return res
}
