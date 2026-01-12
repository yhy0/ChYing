package waf

import (
	"embed"
	"github.com/antlabs/strsim"
	"github.com/chainreactors/proton/templates"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"gopkg.in/yaml.v3"
	"math/rand"
	"strconv"
)

/**
  @author: yhy
  @since: 2023/2/6
  @desc: 提取自 nuclei-templates/technologies/waf-detect.yaml，一个单独的 yaml就不使用 nuclei 测试了
**/

//go:embed waf-detect.yaml
var wafRules embed.FS

var template = &templates.Template{}

var payload = ` AND 1=1 UNION ALL SELECT 1,NULL,'<script>alert(\"XSS\")</script>',table_name FROM information_schema.tables WHERE 2>1--/**/; EXEC xp_cmdshell('cat ../../../etc/passwd')#`

func init() {
	rules, _ := wafRules.ReadFile("waf-detect.yaml")
	err := yaml.Unmarshal(rules, template)
	if err != nil {
		logging.Logger.Errorln(err)
	}
}

func Scan(target, body string, client *httpx.Client) (wafs []string) {
	resp, err := client.Request(target, "POST", "_="+strconv.Itoa(rand.Intn(100))+payload, nil, "Waf-Detect")

	if err != nil {
		logging.Logger.Errorln(err)
		return
	}

	for _, v := range template.RequestsFile[0].Matchers {
		for _, regex := range v.Regex {
			compiled, _ := regexp.Compile(regex)
			if v.Condition == "or" {

			}
			mach := compiled.MatchString(resp.ResponseDump)
			if v.Condition == "or" && mach {
				wafs = append(wafs, v.Name)
				break
			}
		}

	}

	// 如果内置规则没有判断出是否存在 waf，使用页面相似度进行判断，如果相似度小于 0.5 则判断为存在 waf
	if len(wafs) == 0 && strsim.Compare(body, resp.Body) < 0.5 {
		wafs = append(wafs, "unknown(SimilarityRatio)")
	}

	return
}
