package file

import (
	"encoding/json"
	folderutil "github.com/projectdiscovery/utils/folder"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/ChYing/pkg/utils"
	"github.com/yhy0/logging"
	"io/fs"
	"strings"
	"sync"

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
	ChyingDir string

	DictData   []string // DictData 和 JwtSecrets 这两个有点问题，(并发下？，app.go 中的 Startup )通过 fmt.println 这个卡死不动，下面的没有输出，通过 for 循环就可以输出
	JwtSecrets []string

	BBscanRules map[string]*BBscanRule
	Bypass403   map[string][]string
	MitmRules   []MitmRule
	lock        sync.Mutex
)

var (
	regTag           *regexp.Regexp
	regStatus        *regexp.Regexp
	regContentType   *regexp.Regexp
	regContentTypeNo *regexp.Regexp
	regFingerPrints  *regexp.Regexp

	blackText      *regexp.Regexp
	blackRegexText *regexp.Regexp
	blackAllText   *regexp.Regexp
)

func init() {
	// 获取用户配置文件夹的路径
	homedir := folderutil.HomeDirOrDefault("")

	userCfgDir := filepath.Join(homedir, ".config")

	ChyingDir = filepath.Join(userCfgDir, "ChYing")
	// 判断 ChYing 文件夹是否存在
	if _, err := os.Stat(ChyingDir); err != nil {
		// 不存在，创建
		if err = os.MkdirAll(ChyingDir, 0755); err != nil {
			panic(err)
		}
	}

	BBscanRules = make(map[string]*BBscanRule)
	regTag, _ = regexp.Compile(`{tag="(.*?)"}`)
	regStatus, _ = regexp.Compile(`{status=(\d{3})}`)
	regContentType, _ = regexp.Compile(`{type="(.*?)"}`)
	regContentTypeNo, _ = regexp.Compile(`{type_no="(.*?)"}`)
	regFingerPrints, _ = regexp.Compile(`{fingprints="(.*?)"}`)

	blackText, _ = regexp.Compile(`{text="(.*)"}`)
	blackRegexText, _ = regexp.Compile(`{regex_text="(.*)"}`)
	blackAllText, _ = regexp.Compile(`{all_text="(.*)"}`)

	WriteToConfig()
}

func New() {
	// 读取文件
	ReadFiles()

	// 监控配置文件变动
	go watch()
}

// WriteToConfig 将内置文件全部释放到配置文件夹下
func WriteToConfig() {
	// 释放文件内容到 ChYing 文件夹下

	// 1. 释放 dirsearch 扫描规则文件
	content, err := fileDict.ReadFile("dict.txt")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(filepath.Join(ChyingDir, "dict.txt"), content, 0644); err != nil {
		panic(err)
	}

	// 2. 释放 bbscan 规则文件
	bbscan := filepath.Join(ChyingDir, "bbscan")
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
	jwtPath := filepath.Join(ChyingDir, "jwt.txt")
	if !utils.Exists(jwtPath) {
		// 释放 jwt 密钥文件
		jwt, err := jwtSecrets.ReadFile("jwt.txt")
		if err != nil {
			logging.Logger.Fatal(err)
		}
		if err = os.WriteFile(jwtPath, jwt, 0644); err != nil {
			logging.Logger.Fatal(err)
		}
	}

	// 4. 释放 403 bypass 规则文件
	bypass := filepath.Join(ChyingDir, "403bypass")
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

	// 5. 释放 default_mitm_rule.json
	mitmPath := filepath.Join(ChyingDir, "default_mitm_rule.json")
	if !utils.Exists(mitmPath) {
		// 释放 default_mitm_rule.json
		mitmData, err := mitmRules.ReadFile("default_mitm_rule.json")
		if err != nil {
			logging.Logger.Fatal(err)
		}
		if err = os.WriteFile(mitmPath, mitmData, 0644); err != nil {
			logging.Logger.Fatal(err)
		}
	}

}

// ReadFiles 文件存在，读取文件内容
func ReadFiles() {
	lock.Lock()
	defer lock.Unlock()
	dictData, err := os.ReadFile(filepath.Join(ChyingDir, "dict.txt"))
	if err != nil {
		logging.Logger.Errorln("ReadFiles(dict.txt)", err)
		dictData, err = fileDict.ReadFile("dict.txt")
		if err != nil {
			panic(err)
		}
	}
	DictData = strings.Split(string(dictData), "\n")
	BBscanRules = make(map[string]*BBscanRule)
	bbscan := filepath.Join(ChyingDir, "bbscan")
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
		readBBscan(entry.Name(), rulesContent)
	}

	ReadJwtFile("")
	readMitmRule()
	Bypass403 = make(map[string][]string)
	// 返回[]fs.DirEntry
	bypass := filepath.Join(ChyingDir, "403bypass")
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
}

func readBBscan(name string, rulesContent []byte) {
	if name == "black.list" {
		for _, str := range util.CvtLines(string(rulesContent)) {
			if strings.HasPrefix(str, "#") {
				continue
			}
			if !strings.HasPrefix(str, "{") {
				continue
			}
			var black util.BlackRule

			text := blackText.FindStringSubmatch(str)
			if len(text) > 0 {
				black.Type = "text"
				black.Rule = text[1]
				util.BlackLists = append(util.BlackLists, black)
			} else {
				regexText := blackRegexText.FindStringSubmatch(str)
				if len(regexText) > 0 {
					black.Type = "regexText"
					black.Rule = regexText[1]
					util.BlackLists = append(util.BlackLists, black)
				} else {
					allText := blackAllText.FindStringSubmatch(str)
					black.Type = "allText"
					black.Rule = allText[1]
					util.BlackLists = append(util.BlackLists, black)
				}
			}
		}
	} else {
		for _, str := range util.CvtLines(string(rulesContent)) {
			if strings.Index(str, "/") != 0 {
				continue
			}
			var rule BBscanRule

			tag := regTag.FindStringSubmatch(str)
			status := regStatus.FindStringSubmatch(str)
			contentType := regContentType.FindStringSubmatch(str)
			contentTypeNo := regContentTypeNo.FindStringSubmatch(str)
			fingerPrints := regFingerPrints.FindStringSubmatch(str)

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
			if len(fingerPrints) > 0 {
				rule.FingerPrints = strings.Split(fingerPrints[1], ",")
			}

			if util.Contains(str, "{root_only}") {
				rule.Root = true
			}

			path := util.Trim(strings.Split(str, " ")[0])
			BBscanRules[path] = &rule
		}
	}
}

func readMitmRule() {
	mitmData, err := os.ReadFile(filepath.Join(ChyingDir, "default_mitm_rule.json"))
	if err != nil {
		logging.Logger.Errorln("ReadFiles(default_mitm_rule.json)", err)
		mitmData, err = fileDict.ReadFile("default_mitm_rule.json")
		if err != nil {
			panic(err)
		}
	}
	if err := json.Unmarshal(mitmData, &MitmRules); err != nil {
		logging.Logger.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	var _MitmRules []MitmRule
	for _, rule := range MitmRules {
		compiled, _ := regexp.Compile(rule.Rule)
		_MitmRules = append(_MitmRules, MitmRule{
			Rule:              rule.Rule,
			RegexCompiled:     compiled,
			NoReplace:         rule.NoReplace,
			Color:             rule.Color,
			EnableForRequest:  rule.EnableForRequest,
			EnableForResponse: rule.EnableForResponse,
			EnableForHeader:   rule.EnableForHeader,
			EnableForBody:     rule.EnableForBody,
			Index:             rule.Index,
			ExtraTag:          rule.ExtraTag,
			VerboseName:       rule.VerboseName,
		})
	}
	MitmRules = _MitmRules
}

func ReadJwtFile(jwtPath string) {
	if jwtPath == "" {
		jwtPath = filepath.Join(ChyingDir, "jwt.txt")
	}

	jwt, err := os.ReadFile(jwtPath)
	if err != nil {
		logging.Logger.Errorln("ReadFiles(jwt.txt) err:", err)
		jwt, err = jwtSecrets.ReadFile("jwt.txt")
		if err != nil {
			logging.Logger.Fatal(err)
		}
	}
	JwtSecrets = strings.Split(string(jwt), "\n")
}
