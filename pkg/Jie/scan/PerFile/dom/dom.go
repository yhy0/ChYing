package dom

import (
    "fmt"
    regexp "github.com/wasilibs/go-re2"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/logging"
    "strings"
    "time"
)

/**
  @author: yhy
  @since: 2023/3/14
  @desc: https://github1s.com/s0md3v/XSStrike/blob/HEAD/core/dom.py
    // 这种不行，先弃用 没有链路追踪，误报太多
**/

func Dom(u, response string) {
    var highlighted []string
    sources := regexp.MustCompile(`\b(?:document\.(URL|documentURI|URLUnencoded|baseURI|cookie|referrer)|location\.(href|search|hash|pathname)|window\.name|history\.(pushState|replaceState)(local|session)Storage)\b`)
    sinks := regexp.MustCompile(`\b(?:eval|evaluate|execCommand|assign|navigate|getResponseHeaderopen|showModalDialog|Function|set(Timeout|Interval|Immediate)|execScript|crypto.generateCRMFRequest|ScriptElement\.(src|text|textContent|innerText)|.*?\.onEventName|document\.(write|writeln)|.*?\.innerHTML|Range\.createContextualFragment|(document|window)\.location)\b`)
    
    scripts := regexp.MustCompile(`(?i)(?s)<script[^>]*>(.*?)</script>`).FindAllStringSubmatch(response, -1)
    sinkFound, sourceFound := false, false
    for _, script := range scripts {
        lines := strings.Split(script[1], "\n")
        num := 1
        allControlledVariables := make(map[string]bool)
        for _, newLine := range lines {
            line := newLine
            parts := strings.Split(line, "var ")
            
            controlledVariables := make(map[string]bool)
            if len(parts) > 1 {
                for _, part := range parts {
                    for controlledVariable := range allControlledVariables {
                        if strings.Contains(part, controlledVariable) {
                            controlledVariables[regexp.MustCompile(`[a-zA-Z$_][a-zA-Z0-9$_]+`).FindString(part)] = true
                        }
                    }
                }
            }
            pattern := sources.FindAllStringIndex(newLine, -1)
            
            // 寻找 source
            for _, grp := range pattern {
                if grp != nil {
                    source := strings.ReplaceAll(newLine[grp[0]:grp[1]], " ", "")
                    if len(source) > 0 {
                        if len(parts) > 1 {
                            for _, part := range parts {
                                if strings.Contains(part, source) {
                                    controlledVariables[regexp.MustCompile(`[a-zA-Z$_][a-zA-Z0-9$_]+`).FindString(part)] = true
                                }
                            }
                        }
                        line = strings.ReplaceAll(line, source, "*"+source+"*")
                    }
                }
            }
            
            for controlledVariable := range controlledVariables {
                allControlledVariables[controlledVariable] = true
            }
            
            for controlledVariable := range allControlledVariables {
                matches := regexp.MustCompile(`\b`+controlledVariable+`\b`).FindAllStringIndex(line, -1)
                if len(matches) > 0 {
                    sourceFound = true
                    line = regexp.MustCompile(`\b`+controlledVariable+`\b`).ReplaceAllString(line, "**"+controlledVariable+"**")
                }
            }
            
            // 寻找 sink
            pattern = sinks.FindAllStringIndex(newLine, -1)
            
            for _, grp := range pattern {
                if grp != nil {
                    sink := strings.ReplaceAll(newLine[grp[0]:grp[1]], " ", "")
                    if len(sink) > 0 {
                        line = strings.ReplaceAll(line, sink, "*"+sink+"*")
                        sinkFound = true
                    }
                }
            }
            if line != newLine {
                highlighted = append(highlighted, fmt.Sprintf("%-3d %s", num, strings.TrimLeft(line, " ")))
            }
            num += 1
        }
    }
    
    if sinkFound || sourceFound {
        output.OutChannel <- output.VulMessage{
            DataType: "web_vul",
            Plugin:   "XSS",
            VulnData: output.VulnData{
                CreateTime: time.Now().Format("2006-01-02 15:04:05"),
                Target:     u,
                VulnType:   "Dom1 XSS",
                Payload:    strings.Join(highlighted, "\t"),
            },
            Level: output.Medium,
        }
        logging.Logger.Infoln(u, highlighted)
    }
}
