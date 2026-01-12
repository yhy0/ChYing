package sensitive

import (
    "embed"
    regexp "github.com/wasilibs/go-re2"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "gopkg.in/yaml.v3"
    "io/fs"
    "strings"
    "time"
)

/**
   @author yhy
   @since 2024/4/16
   @desc https://github.com/1c3z/arl_files/tree/master/wih
**/

//go:embed wih.yml
var wih embed.FS

// Rule 结构体代表YAML文件中的一条规则
type Rule struct {
    ID      string `yaml:"id"`
    Enabled bool   `yaml:"enabled"`
    Pattern string `yaml:"pattern"`
}

// Rules 结构体包含多个Rule
type Rules struct {
    Rules []Rule `yaml:"rules"`
}

// 预编译正则表达式, 避免每次匹配都需要编译的开销。
var wihRegexCompiled map[string]*regexp.Regexp

func init() {
    // 读取嵌入的文件内容
    data, _ := fs.ReadFile(wih, "wih.yml")
    var wihRules Rules
    yaml.Unmarshal(data, &wihRules)
    
    wihRegexCompiled = make(map[string]*regexp.Regexp, len(wihRules.Rules))
    for _, rule := range wihRules.Rules {
        compiled, _ := regexp.Compile(rule.Pattern)
        wihRegexCompiled[rule.ID] = compiled
    }
}

func Wih(url, req, body string) bool {
    flag := false
    for name, regex := range wihRegexCompiled {
        var matchedRegexes []string
        
        matches := regex.FindAllString(body, -1)
        for _, m := range matches {
            mm := strings.Split(m, ":")
            // Token:    去除
            if len(mm) > 1 && strings.Trim(mm[1], " ") == "" {
                continue
            }
            
            mm = strings.Split(m, "=")
            // Token=    去除
            if len(mm) > 1 && strings.Trim(mm[1], " ") == "" {
                continue
            }
            matchedRegexes = append(matchedRegexes, m)
        }
        
        if len(matchedRegexes) > 0 {
            flag = true
            output.OutChannel <- output.VulMessage{
                DataType: "web_vul",
                Plugin:   "Sensitive Key",
                VulnData: output.VulnData{
                    VulnType:   name,
                    CreateTime: time.Now().Format("2006-01-02 15:04:05"),
                    Target:     url,
                    Payload:    strings.Join(matchedRegexes, ","),
                },
                Level: output.Medium,
            }
        }
    }
    
    return flag
}
