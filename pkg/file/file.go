package file

import (
	"encoding/json"
	folderutil "github.com/projectdiscovery/utils/folder"
	"github.com/yhy0/Jie/pkg/util"
	"github.com/yhy0/logging"
	"io/fs"
	"regexp"
	"strings"

	"os"
	"path/filepath"
)

/**
  @author: yhy
  @since: 2023/4/21
  @desc: 处理内置文件释放到 $user/.config/ChYing 文件夹下
	默认会使用配置文件下的，这样更改扫描规则时就不用重新打包了
**/

var (
	DiccData    []string
	BBscanRules map[string]*Rule
	JwtSecrets  []string
	Bypass403   map[string][]string
	Av          map[string]string
)

var chyingDir string

type Rule struct {
	Tag    string // 文本内容
	Status string // 状态码
	Type   string // 返回的 ContentType
	TypeNo string // 不可能返回的 ContentType
	Root   bool   // 是否为一级目录
}

func New() {
	diccData, err := fileDicc.ReadFile("dicc.txt")
	if err != nil {
		panic(err)
	}

	DiccData = strings.Split(string(diccData), "\n")

	jwt, err := jwtSecrets.ReadFile("twj.txt")
	if err != nil {
		panic(err)
	}

	JwtSecrets = strings.Split(string(jwt), "\n")

	BBscanRules = make(map[string]*Rule)
	regTag, _ := regexp.Compile(`{tag="(.*?)"}`)
	regStatus, _ := regexp.Compile(`{status=(\d{3})}`)
	regContentType, _ := regexp.Compile(`{type="(.*?)"}`)
	regContentTypeNo, _ := regexp.Compile(`{type_no="(.*?)"}`)

	// 返回[]fs.DirEntry
	entries, err := bbscanRules.ReadDir("bbscan")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		rulesContent, err := bbscanRules.ReadFile("bbscan/" + entry.Name())
		if err != nil {
			continue
		}

		for _, str := range util.CvtLines(string(rulesContent)) {
			if strings.Index(str, "/") != 0 {
				continue
			}
			var rule = &Rule{}

			tag := regTag.FindStringSubmatch(str)
			status := regStatus.FindStringSubmatch(str)
			contentType := regContentType.FindStringSubmatch(str)
			contentTypeNo := regContentTypeNo.FindStringSubmatch(str)

			if len(tag) > 0 {
				rule.Tag = tag[1]
			}

			if len(status) > 0 {
				rule.Status = status[1]
			}
			if len(contentType) > 0 {
				rule.Type = contentType[1]
			}
			if len(contentTypeNo) > 0 {
				rule.TypeNo = contentTypeNo[1]
			}

			if util.Contains(str, "{root_only}") {
				rule.Root = true
			}
			path := Trim(strings.Split(str, " ")[0])
			BBscanRules[path] = rule
		}
	}

	Bypass403 = make(map[string][]string)
	// 返回[]fs.DirEntry
	entries, err = bypass403.ReadDir("403bypass")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		content, err := bypass403.ReadFile("403bypass/" + entry.Name())
		if err != nil {
			continue
		}
		Bypass403[entry.Name()] = util.CvtLines(string(content))
	}

	// 读取 av.json 文件内容
	data, err := AvFile.ReadFile("av.json")
	if err != nil {
		panic(err)
	}
	// 解析 JSON 数据到一个 map 对象中
	Av = make(map[string]string)
	err = json.Unmarshal(data, &Av)
	if err != nil {
		panic(err)
	}

	// 将文件释放
	UserFile()
}

func UserFile() {
	// 获取用户配置文件夹的路径
	homedir := folderutil.HomeDirOrDefault("")

	userCfgDir := filepath.Join(homedir, ".config")

	filePath := filepath.Join(userCfgDir, "ChYing")

	// 判断 ChYing 文件夹是否存在
	if _, err := os.Stat(filePath); err != nil {
		// 不存在，创建
		chyingDir = filepath.Join(userCfgDir, "ChYing")
		if err := os.MkdirAll(chyingDir, 0755); err != nil {
			panic(err)
		}
		WriteToConfig()
	} else {
		chyingDir = filepath.Join(userCfgDir, "ChYing")
		// 读取文件
		ReadFiles()
	}
}

// WriteToConfig 将内置文件全部释放到配置文件夹下
func WriteToConfig() {
	// 释放文件内容到 ChYing 文件夹下
	// 1. 释放 dirsearch 扫描规则文件
	content, err := fileDicc.ReadFile("dicc.txt")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(chyingDir, "dicc.txt"), content, 0644); err != nil {
		panic(err)
	}

	// 2. 释放 bbscan 规则文件
	bbscan := filepath.Join(chyingDir, "bbscan")
	if err := os.MkdirAll(bbscan, 0755); err != nil {
		panic(err)
	}
	// 遍历嵌入文件夹中的文件
	files, err := fs.ReadDir(bbscanRules, "bbscan")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		// 拼接目标文件路径
		ruleFile := filepath.Join(bbscan, f.Name())

		rulesContent, err := bbscanRules.ReadFile("bbscan/" + f.Name())
		if err != nil {
			logging.Logger.Errorln(err)
			continue
		}

		if err = os.WriteFile(ruleFile, rulesContent, 0644); err != nil {
			logging.Logger.Errorln(err)
			continue
		}
	}

	// 3. 释放 jwt 密钥文件
	jwt, err := jwtSecrets.ReadFile("twj.txt")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(chyingDir, "twj.txt"), jwt, 0644); err != nil {
		panic(err)
	}

	// 4. 释放 403 bypass 规则文件
	bypass := filepath.Join(chyingDir, "403bypass")
	if err = os.MkdirAll(bypass, 0755); err != nil {
		panic(err)
	}
	// 遍历嵌入文件夹中的文件
	files, err = fs.ReadDir(bypass403, "403bypass")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		// 拼接目标文件路径
		ruleFile := filepath.Join(bypass, f.Name())

		rulesContent, err := bypass403.ReadFile("403bypass/" + f.Name())
		if err != nil {
			logging.Logger.Errorln(err)
			continue
		}

		if err = os.WriteFile(ruleFile, rulesContent, 0644); err != nil {
			logging.Logger.Errorln(err)
			continue
		}
	}

	// 5. 释放杀软对照文件
	av, err := AvFile.ReadFile("av.json")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(chyingDir, "av.json"), av, 0644); err != nil {
		panic(err)
	}
}

// ReadFiles 文件存在，读取文件内容
func ReadFiles() {
	diccData, err := os.ReadFile(filepath.Join(chyingDir, "dicc.txt"))
	if err != nil {
		logging.Logger.Errorln("ReadFiles(dicc.txt)", err)
		diccData, err = fileDicc.ReadFile("dicc.txt")
		if err != nil {
			panic(err)
		}
	}
	DiccData = strings.Split(string(diccData), "\n")

	jwt, err := os.ReadFile(filepath.Join(chyingDir, "twj.txt"))
	if err != nil {
		logging.Logger.Errorln("ReadFiles(twj.txt)", err)
		diccData, err = jwtSecrets.ReadFile("twj.txt")
		if err != nil {
			panic(err)
		}
	}
	JwtSecrets = strings.Split(string(jwt), "\n")

	BBscanRules = make(map[string]*Rule)
	regTag, _ := regexp.Compile(`{tag="(.*?)"}`)
	regStatus, _ := regexp.Compile(`{status=(\d{3})}`)
	regContentType, _ := regexp.Compile(`{type="(.*?)"}`)
	regContentTypeNo, _ := regexp.Compile(`{type_no="(.*?)"}`)

	bbscan := filepath.Join(chyingDir, "bbscan")
	entries, err := os.ReadDir(bbscan)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		rulesContent, err := os.ReadFile(bbscan + "/" + entry.Name())
		if err != nil {
			logging.Logger.Errorf("ReadFiles(bbscan/%s): %v", entry.Name(), err)
			continue
		}

		for _, str := range util.CvtLines(string(rulesContent)) {
			if strings.Index(str, "/") != 0 {
				continue
			}
			var rule = &Rule{}

			tag := regTag.FindStringSubmatch(str)
			status := regStatus.FindStringSubmatch(str)
			contentType := regContentType.FindStringSubmatch(str)
			contentTypeNo := regContentTypeNo.FindStringSubmatch(str)

			if len(tag) > 0 {
				rule.Tag = tag[1]
			}

			if len(status) > 0 {
				rule.Status = status[1]
			}
			if len(contentType) > 0 {
				rule.Type = contentType[1]
			}
			if len(contentTypeNo) > 0 {
				rule.TypeNo = contentTypeNo[1]
			}

			if util.Contains(str, "{root_only}") {
				rule.Root = true
			}
			path := Trim(strings.Split(str, " ")[0])
			BBscanRules[path] = rule
		}
	}

	Bypass403 = make(map[string][]string)
	// 返回[]fs.DirEntry
	bypass := filepath.Join(chyingDir, "403bypass")
	entries, err = os.ReadDir(bypass)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		content, err := os.ReadFile(bypass + "/" + entry.Name())
		if err != nil {
			continue
		}
		Bypass403[entry.Name()] = util.CvtLines(string(content))
	}

	avfile := filepath.Join(chyingDir, "av.json")
	_, err = os.Stat(avfile)
	if os.IsNotExist(err) {
		av, err := AvFile.ReadFile("av.json")
		if err != nil {
			panic(err)
		}
		if err = os.WriteFile(avfile, av, 0644); err != nil {
			panic(err)
		}
	} else {
		// 读取 av.json 文件内容
		data, err := os.ReadFile(filepath.Join(chyingDir, "av.json"))
		if err != nil {
			logging.Logger.Errorln("ReadFiles(twj.txt)", err)
			data, err = AvFile.ReadFile("av.json")
			if err != nil {
				panic(err)
			}
		}
		// 解析 JSON 数据到一个 map 对象中
		Av = make(map[string]string)
		err = json.Unmarshal(data, &Av)
		if err != nil {
			panic(err)
		}
	}
}

func Trim(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}
