package sensitive

import (
    "embed"
    "github.com/chainreactors/proton/templates"
    regexp "github.com/wasilibs/go-re2"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "gopkg.in/yaml.v3"
    "io/fs"
    "strings"
    "time"
)

/**
  @author: yhy
  @since: 2022/7/22
  @desc: 提取 https://github.com/projectdiscovery/nuclei-templates/tree/main/file/keys 中的规则
**/

//go:embed keys/*
var ruleFiles embed.FS
var rules []templates.Template

// 限制一下正则执行，防止执行过多导致 CPU 爆炸
var limit chan struct{}

// 预编译正则表达式, 避免每次匹配都需要编译的开销。
var regexCompiled map[string][]*regexp.Regexp

// 加载规则
func init() {
    // 读取规则文件
    ruleDir, err := fs.ReadDir(ruleFiles, "rules")
    if err != nil {
        return
    }
    
    for _, file := range ruleDir {
        if file.IsDir() {
            ruleDir2, _ := fs.ReadDir(ruleFiles, "rules/"+file.Name())
            
            for _, file2 := range ruleDir2 {
                content, err := ruleFiles.ReadFile("rules/" + file.Name() + "/" + file2.Name())
                if err != nil {
                    continue
                }
                
                var rule templates.Template
                err = yaml.Unmarshal(content, &rule)
                if err != nil || rule.Id == "" {
                    continue
                }
                rules = append(rules, rule)
            }
        }
        
        content, _ := ruleFiles.ReadFile("rules/" + file.Name())
        
        var rule templates.Template
        err := yaml.Unmarshal(content, &rule)
        if err != nil || rule.Id == "" {
            continue
        }
        
        rules = append(rules, rule)
    }
    
    // 预编译正则
    regexCompiled = make(map[string][]*regexp.Regexp, len(rules))
    for _, rule := range rules {
        e := rule.RequestsFile[0].Operators.Extractors[0]
        
        for _, regex := range e.Regex {
            compiled, err := regexp.Compile(regex)
            if err != nil {
                continue
            }
            regexCompiled[rule.Id] = append(regexCompiled[rule.Id], compiled)
        }
        
    }
    // todo 800 好像还是有点多，400 吧
    limit = make(chan struct{}, 400)
    
}

// KeyDetection 页面敏感信息检测
func KeyDetection(url, body string) bool {
    flag := false
    for id, regexs := range regexCompiled {
        var matchedRegexes []string
        for _, regex := range regexs {
            limit <- struct{}{}
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
            <-limit
        }
        
        if len(matchedRegexes) > 0 {
            output.OutChannel <- output.VulMessage{
                DataType: "web_vul",
                Plugin:   "Sensitive Key",
                VulnData: output.VulnData{
                    VulnType:   id,
                    CreateTime: time.Now().Format("2006-01-02 15:04:05"),
                    Target:     url,
                    Payload:    strings.Join(matchedRegexes, ","),
                    // Response:   body, // todo js 这种文本过大，不显示了
                },
                Level: output.Medium,
            }
            flag = true
            
        }
    }
    
    return flag
}
