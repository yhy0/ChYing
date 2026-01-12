package pocs_go

import (
    "fmt"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/shiro"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/weblogic"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/oa/seeyon"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/oa/yongyou/nc"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/pocs_go/ThinkPHP"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/pocs_go/jboss"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/pocs_go/jenkins"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/pocs_go/phpunit"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/pocs_go/tomcat"
    "github.com/yhy0/ChYing/pkg/Jie/scan/gadget/brute"
    "net/url"
    "strings"
    "time"
)

func PocCheck(technologies []string, target string, finalURL string, ip string, check map[string]bool, client *httpx.Client) map[string]bool {
    vulnerability := false
    var (
        plugin  string
        payload string
    )
    
    for _, wt := range technologies {
        wt = strings.ToLower(wt)
        if strings.Contains(wt, "shiro") && check["shiro"] == false {
            check["shiro"] = true
            key, mode := shiro.CVE_2016_4437(finalURL, "", client)
            if key != "" {
                vulnerability = true
                plugin = "Shiro"
                payload = mode + ": " + key
            }
        } else if strings.Contains(wt, "tomcat") && check["tomcat"] == false {
            check["tomcat"] = true
            username, password := brute.TomcatBrute(target, client)
            if username != "" {
                vulnerability = true
                plugin = "Apache Tomcat"
                payload = fmt.Sprintf("brute-Tomcat|%s:%s", username, password)
            }
            var HOST string
            if host, err := url.Parse(target); err == nil {
                HOST = host.Host
            }
            
            if tomcat.CVE_2020_1938(HOST) {
                vulnerability = true
                plugin = "Apache Tomcat"
                payload += "exp-Tomcat|CVE_2020_1938 \n"
            }
            if tomcat.CVE_2017_12615(target, client) {
                vulnerability = true
                plugin = "Apache Tomcat"
                payload += "exp-Tomcat|CVE_2017_12615 \n"
            }
        } else if strings.Contains(wt, "basic") && check["basic"] == false { // todo 这里还没有匹配到
            check["basic"] = true
            username, password, _ := brute.BasicBrute(target, client)
            if username != "" {
                vulnerability = true
                plugin = "Basic"
                payload = fmt.Sprintf("brute-basic|%s:%s", username, password)
            }
        } else if strings.Contains(wt, "weblogic") && check["WebLogic"] == false {
            check["WebLogic"] = true
            username, password := brute.WeblogicBrute(target, client)
            if username != "" {
                vulnerability = true
                plugin = "WebLogic"
                payload = fmt.Sprintf("brute-Weblogic|%s:%s \n", username, password)
            }
            if weblogic.CVE_2014_4210(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2014_4210 \n"
            }
            if weblogic.CVE_2017_3506(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2017_3506 \n"
            }
            if weblogic.CVE_2017_10271(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2017_10271 \n"
            }
            if weblogic.CVE_2018_2894(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2018_2894 \n"
            }
            if weblogic.CVE_2019_2725(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2019_2725 \n"
            }
            if weblogic.CVE_2019_2729(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2019_2729 \n"
            }
            if weblogic.CVE_2020_2883(target) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2020_2883 \n"
            }
            if weblogic.CVE_2020_14882(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2020_14882 \n"
            }
            if weblogic.CVE_2020_14883(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2020_14883 \n"
            }
            if weblogic.CVE_2021_2109(target, client) {
                vulnerability = true
                plugin = "WebLogic"
                payload += "exp-WebLogic|CVE_2021_2109\n"
            }
        } else if strings.Contains(wt, "jboss") && check["Jboss"] == false {
            check["Jboss"] = true
            if jboss.CVE_2017_12149(target, client) {
                vulnerability = true
                plugin = "Jboss"
                payload += "exp-Jboss|CVE_2017_12149| \n"
            }
            username, password := brute.JbossBrute(target, client)
            if username != "" {
                vulnerability = true
                plugin = "Jboss"
                payload += fmt.Sprintf("brute-Jboss|%s:%s", username, password)
            }
        } else if strings.Contains(wt, "jenkins") && check["Jenkins"] == false {
            check["Jenkins"] = true
            if jenkins.Unauthorized(target, client) {
                vulnerability = true
                plugin = "Jenkins"
                payload += "exp-Jenkins|Unauthorized script \n"
            }
            if jenkins.CVE_2018_1000110(target, client) {
                vulnerability = true
                plugin = "Jenkins"
                payload += "exp-Jenkins|CVE_2018_1000110 \n"
            }
            if jenkins.CVE_2018_1000861(target, client) {
                vulnerability = true
                plugin = "Jenkins"
                payload += "exp-Jenkins|CVE_2018_1000861 \n"
            }
            if jenkins.CVE_2019_10003000(target, client) {
                vulnerability = true
                plugin = "Jenkins"
                payload += "exp-Jenkins|CVE_2019_10003000 \n"
            }
        } else if strings.Contains(wt, "thinkphp") && check["ThinkPHP"] == false {
            check["ThinkPHP"] = true
            if ThinkPHP.RCE(target, client) {
                vulnerability = true
                plugin = "ThinkPHP"
                payload = "exp-ThinkPHP \n"
            }
        } else if strings.Contains(wt, "phpunit") && check["phpunit"] == false {
            check["phpunit"] = true
            if phpunit.CVE_2017_9841(target, client) {
                vulnerability = true
                plugin = "phpunit"
                payload = "exp-phpunit|CVE_2017_9841 \n"
            }
        } else if strings.Contains(wt, "seeyon") && check["seeyon"] == false {
            check["seeyon"] = true
            if seeyon.SeeyonFastjson(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|SeeyonFastjson \n"
            }
            if seeyon.SessionUpload(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|SessionUpload \n"
            }
            if seeyon.CNVD_2019_19299(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|CNVD_2019_19299 \n"
            }
            if seeyon.CNVD_2020_62422(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|CNVD_2020_62422 \n"
            }
            if seeyon.CNVD_2021_01627(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|CNVD_2021_01627 \n"
            }
            if seeyon.CreateMysql(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|CreateMysql \n"
            }
            if seeyon.DownExcelBeanServlet(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|DownExcelBeanServlet \n"
            }
            if seeyon.GetSessionList(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|GetSessionList \n"
            }
            if seeyon.InitDataAssess(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|InitDataAssess \n"
            }
            if seeyon.ManagementStatus(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|ManagementStatus \n"
            }
            if seeyon.BackdoorScan(target, client) {
                vulnerability = true
                plugin = "seeyon"
                payload += "exp-seeyon|Backdoor \n"
            }
        } else if (strings.Contains(wt, "loginPage") || strings.Contains(wt, "登录")) && check["loginPage"] == false {
            check["loginPage"] = true
            username, password, loginurl := brute.Admin_brute(finalURL, client)
            if loginurl != "" {
                vulnerability = true
                plugin = "LoginPage"
                payload += fmt.Sprintf("brute-admin|%s:%s", username, password)
            }
        } else if strings.Contains(wt, "用友NC") && check["YongYouNc"] == false {
            check["YongYouNc"] = true
            if nc.Scan(target, client) {
                vulnerability = true
                plugin = "用友 NC"
                payload = "用友 NC|反序列化"
            }
            // } else if strings.Contains(wt, "shiro") {
            // case "sunlogin":
            //    if sunlogin.SunloginRCE(target) {
            //        technologies = append(technologies, "exp-Sunlogin|RCE")
            //    }
            // case "zabbixsaml":
            //    if zabbix.CVE_2022_23131(target) {
            //        technologies = append(technologies, "exp-ZabbixSAML|bypass-login")
            //    }
            // case "spring", "spring env", "spring-boot", "spring-framework", "spring-boot-admin":
            //    if Springboot.CVE_2022_22965(finalURL) {
            //        technologies = append(technologies, "exp-Spring4Shell|CVE_2022_22965")
            //    }
            // case "springgateway":
            //    if Springboot.CVE_2022_22947(target) {
            //        technologies = append(technologies, "exp-SpringGateway|CVE_2022_22947")
            //    }
            // case "gitLab":
            //    if gitlab.CVE_2021_22205(target) {
            //        technologies = append(technologies, "exp-gitlab|CVE_2021_22205")
            //    }
        }
    }
    
    if vulnerability {
        output.OutChannel <- output.VulMessage{
            DataType: "web_vul",
            Plugin:   plugin,
            VulnData: output.VulnData{
                CreateTime: time.Now().Format("2006-01-02 15:04:05"),
                Target:     target,
                Ip:         ip,
                Payload:    payload,
            },
            Level: output.Critical,
        }
    }
    
    return check
}
