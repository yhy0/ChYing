package nucleiY

import (
	"fmt"
	"github.com/projectdiscovery/nuclei/v2/pkg/core/inputs"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/contextargs"
	"github.com/projectdiscovery/nuclei/v2/pkg/templates"
	"github.com/thoas/go-funk"
	"strings"
)

/**
   @author yhy
   @since 2023/6/8
   @desc 基于 nuclei 实现的重点漏洞扫描
**/

type Info struct {
	Name     string                `json:"name"`
	Template []*templates.Template `json:"template"`
}

type Options struct {
	Label    string   `json:"label"`
	Children []string `json:"children"`
}

type Result struct {
	Url      string `json:"url"`
	Name     string `json:"name"`
	Request  string `json:"request"`
	Response string `json:"response"`
}

var Pocs map[string][]*templates.Template

func Scan(target string, tag string, proxy string) string {
	if nuclei == nil {
		return "nuclei == nil"
	}
	var ts []*templates.Template

	if tag == "all" { // 使用全部 poc 探测
		ts = nuclei.Store.Templates()
	} else if funk.Contains(tag, "-all") { // 使用每个分类的全部 poc 探测
		tag = strings.ReplaceAll(tag, "-all", "")
		value, ok := Pocs[tag]
		if ok {
			ts = value
		} else {
			return fmt.Sprintf("tag[%s] not find", tag)
		}
	} else { // 单个探测
		_tag := strings.Split(tag, ":")
		for _, t := range Pocs[_tag[0]] {
			if t.Info.Name == _tag[1] {
				ts = append(ts, t)
				break
			}
		}
	}

	if len(ts) > 0 {
		input := &inputs.SimpleInputProvider{Inputs: []*contextargs.MetaInput{{Input: target}}}
		_ = nuclei.Engine.Execute(ts, input)
		nuclei.Engine.WorkPool().Wait() // Wait for the scan to finish
		return ""
	} else {
		return "没有模板"
	}
}
