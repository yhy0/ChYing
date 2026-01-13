package collection

import (
    "github.com/thoas/go-funk"
    regexp "github.com/wasilibs/go-re2"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/util"
    "github.com/yhy0/logging"
    "golang.org/x/net/publicsuffix"
    "net"
    "net/url"
    "strings"
)

/**
  @author: yhy
  @since: 2023/11/1
  @desc: //TODO
**/

// Info  domain: 用来限制获取的子域名
func Info(target, domain string, body string, contentType string) (c output.Collection) {
    logging.Logger.Debugln("start collection url:", target)
    var domains []string
    for _, v := range conf.GlobalConfig.Collection.Domain {
        re := regexp.MustCompile(v)
        domains = util.RemoveQuotation(re.FindAllString(body, -1))
    }
    
    // 使用 publicsuffix 包获取二级域名
    _domain, _ := publicsuffix.EffectiveTLDPlusOne(domain)
    for _, d := range domains {
        // 正则会匹配到 .com.cn 这种，需要过滤掉 . 开头的
        if strings.HasPrefix(d, ".") {
            continue
        }
        d = strings.ReplaceAll(d, "http://", "")
        d = strings.ReplaceAll(d, "https://", "")
        d = strings.ReplaceAll(d, "://", "")
        d = strings.ReplaceAll(d, "//", "")
        if strings.Contains(d, _domain) {
            c.Subdomain = append(c.Subdomain, d)
        } else {
            c.OtherDomain = append(c.OtherDomain, d)
        }
    }
    
    var ips []string
    for _, v := range conf.GlobalConfig.Collection.IP {
        re := regexp.MustCompile(v)
        ips = util.RemoveQuotation(re.FindAllString(body, -1))
    }
    for _, i := range ips {
        // 正则会匹配到 .com.cn 这种，需要过滤掉 . 开头的
        if strings.HasPrefix(i, ".") {
            continue
        }
        i = strings.ReplaceAll(i, "http://", "")
        i = strings.ReplaceAll(i, "https://", "")
        i = strings.ReplaceAll(i, "://", "")
        i = strings.ReplaceAll(i, "//", "")
        // 不带端口号的，需要验证一下是否为 ip ，目前正则会匹配到 1.2.840.100 这种
        if !strings.Contains(i, ":") && net.ParseIP(i) == nil {
            continue
        }
        if util.IsInnerIP(i) {
            c.InnerIp = append(c.InnerIp, i)
        } else {
            c.PublicIp = append(c.PublicIp, i)
        }
    }
    for _, v := range conf.GlobalConfig.Collection.Phone {
        re := regexp.MustCompile(v)
        c.Phone = append(c.Phone, util.RemoveQuotation(re.FindAllString(body, -1))...)
    }
    
    for _, v := range conf.GlobalConfig.Collection.Email {
        re := regexp.MustCompile(v)
        c.Email = append(c.Email, util.RemoveQuotation(re.FindAllString(body, -1))...)
    }
    
    for _, v := range conf.GlobalConfig.Collection.IDCard {
        re := regexp.MustCompile(v)
        c.IdCard = append(c.IdCard, util.RemoveQuotation(re.FindAllString(body, -1))...)
    }
    
    for _, v := range conf.GlobalConfig.Collection.Other {
        re := regexp.MustCompile(v)
        c.Others = append(c.Others, util.RemoveQuotation(re.FindAllString(body, -1))...)
    }
    
    c.Api = Api(target, body, contentType)
    
    for _, v := range conf.GlobalConfig.Collection.Url {
        re := regexp.MustCompile(v)
        urls := re.FindAllStringSubmatch(body, -1)
        urls = urlFilter(urls)
        // 循环提取url放到结果中
        for _, u := range urls {
            if u[0] == "" {
                continue
            }
            c.Urls = append(c.Urls, u[0])
        }
    }
    return
}

func Api(target, body string, contentType string) []string {
    var res []string
    for _, v := range conf.GlobalConfig.Collection.API {
        re := regexp.MustCompile(v)
        apis := re.FindAllStringSubmatch(body, -1)
        for _, u := range apis {
            _u := util.RemoveQuotationMarks(u[0])
            // 正则识别出来的有空格、<、> 的排除，基本都是误报
            if len(u) > 1 && (strings.Contains(_u, " ") || strings.Contains(_u, "<") || strings.Contains(_u, ">")) {
                continue
            }
            
            if strings.HasSuffix(_u, ".css") || strings.HasSuffix(_u, ".js") || strings.Contains(_u, ".js?") || strings.Contains(_u, ".css?") || _u == "/a/b" {
                continue
            }
            
            if len(u) < 3 {
                // "(?:\"|')(/[^/\"']+){2,}(?:\"|')"
                if _u == "" || !strings.HasPrefix(_u, "/") {
                    continue
                }
                res = append(res, _u)
            } else {
                // "(?i)\\.(get|post|put|delete|options|connect|trace|patch)\\([\"'](/?.*?)[\"']" 这个正则
                // 不是以 / 开头的去除
                if u[2] == "" || !strings.HasPrefix(u[2], "/") {
                    continue
                }
                res = append(res, u[2])
            }
            logging.Logger.Debugln(target, u)
        }
    }
    
    if funk.Contains(contentType, "application/javascript") {
        res = append(res, analyzeJsluice(target, body)...)
    }
    return funk.UniqString(res)
}

func urlFilter(str [][]string) [][]string {
    // 对不需要的数据过滤
    for i := range str {
        _str := strings.Join(str[i], "")
        if strings.Contains(_str, "YYYY/MM") || strings.Contains(_str, "MM/YYYY") || strings.Contains(_str, "YYYY-MM") || strings.Contains(_str, "MM-YYYY") {
            continue
        }
        
        if strings.Contains(_str, "images/png") {
            continue
        }
        
        if len(str[i]) > 1 {
            str[i][0], _ = url.QueryUnescape(str[i][1])
        }
        str[i][0] = strings.TrimSpace(str[i][0])
        str[i][0] = strings.Replace(str[i][0], " ", "", -1)
        str[i][0] = strings.Replace(str[i][0], "\\/", "/", -1)
        str[i][0] = strings.Replace(str[i][0], "%3A", ":", -1)
        str[i][0] = strings.Replace(str[i][0], "%2F", "/", -1)
        // 去除不存在字符串和数字的url,判断为错误数据
        match, _ := regexp.MatchString("[a-zA-Z]+|[0-9]+", str[i][0])
        if !match {
            str[i][0] = ""
            continue
        }
        
        // 对抓到的域名做处理
        re := regexp.MustCompile("([a-z0-9\\-]+\\.)+([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?").FindAllString(str[i][0], 1)
        if len(re) != 0 && !strings.HasPrefix(str[i][0], "http") && !strings.HasPrefix(str[i][0], "/") {
            str[i][0] = "http://" + str[i][0]
        }
        
        // 过滤配置的黑名单
        for i2 := range conf.GlobalConfig.Collection.UrlFilter {
            _re := regexp.MustCompile(conf.GlobalConfig.Collection.UrlFilter[i2])
            is := _re.MatchString(str[i][0])
            if is {
                str[i][0] = ""
                break
            }
        }
        
    }
    return str
}
