package conf

import (
    "fmt"
    "os"
    "os/exec"
)

/**
  @author: yhy
  @since: 2024/4/12
  @desc: 检查 nmap、masscan、chrome 是否已经安装
**/

func Preparations() {
    // if !GlobalConfig.NoPortScan { // 不进行端口扫描时，不检查这些
    Plugin["portScan"] = false
    // 检查 nmap 是否已安装
    nmapInstalled := commandExists("nmap")
    if !nmapInstalled {
        fmt.Println("nmap not found, please install")
        os.Exit(1)
    }
    
    // 检查 masscan 是否已安装
    masscanInstalled := commandExists("masscan")
    if !masscanInstalled {
        fmt.Println("masscan not found, please install")
        os.Exit(1)
    }
    // }
}

// 检查命令是否可执行
func commandExists(cmd string) bool {
    _, err := exec.LookPath(cmd)
    return err == nil
}
