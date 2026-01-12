package fuzz

import (
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/util"
    "strings"
)

/**
  @author: yhy
  @since: 2023/4/23
  @desc: //TODO
**/

var Page404Title = []string{"404", "不存在", "错误", "403", "访问禁止", "禁止访问", "请求含有不合法的参数", "无法访问", "云防护", "网络防火墙", "网站防火墙", "访问拦截", "由于安全原因JSP功能默认关闭", "Site Not Found Exception", "AccessDeny", "502 Bad Gateway", "Bad Request", "illegal URL", "出错了", "504 Gateway", "Internal Server Error"}
var Page404Content = []string{"系统错误", "https://imgcache.qq.com/qcloud/security/static/404style.css", "<script>document.getElementById(\"a-link\").click();</script>", "404 Not Found", "您所提交的请求含有不合法的参数，已被网站管理员设置拦截", "404.safedog.cn", "URL was rejected", "hello, are you lost?", "没有找到站点:", "\"error_msg\":\"参数错误", "document.location='/host_not_found_error';", "页面不存在", "页面没有找到", "访问禁止", "方法不存在", "方法错误", "\"code\":404", "\"code\":\"404", "\"status\":404,", "This page can't be displayed. Contact support for additional information", "not supported url path", "assets.alicdn.com/g/dt/tracker/4.0.0/??tracker.Tracker.js", "tracker.Tracker.js", "bixi.alicdn.com/punish/punish:resource:template:", "请求路径无效", "页面走丢了", "资源不存在", "\"code\":\"130001\",", "\\u672a\\u77e5\\u9519\\u8bef", "The following error was encountered:", "系统出现错误", "An error occurred.", "\"error\":\"invalid url\"", "501 Not Implemented", "a padding to disable MSIE and Chrome friendly error page", "服务异常，请联系管理员", "\"error\":\"invalid out format\"", "document.title=\"出错了\";", "document.title=\"Error\";", "The specified bucket does not exist", "The specified bucket is not valid", "页面无法访问", "无效链接", "\"msg\":\"Not Found\"", "Request Not Authorized"}
var Page403title = []string{"403", "Forbidden", "ERROR", "error", "Bad Message"}
var Page403Content = []string{"Unsupported openapi method", "Forbidden", "AccessDenied", "Access Denied", "Access defined", "token is invalid", "请求非法", "禁止访问", "未认证", "认证不通过", "认证失败", "未登录", "请登录", "请先登录", "鉴权失败", "登录失败", "登录凭证无效", "您无权访问该资源,请联系系统管理员", "batit.aliyun.com/alww.html", "You don't have permission to access", "重新登录", "need login", "\"code\":401", "\"code\":403", "非法IP调用", "阻断该请求", "登录信息超时", "message\":\"un login", "PERMISSION_DENIED", "非法请求", "Authorization header", "not support this", "\":500,", "\":403,", "\":404,", "非法访问", "you are not permitted to", "needLogin"}
var Location404 = []string{"/auth/login/", "error.html", "User timed out!", "[session timeout]"}
var WafContent = []string{"g.alicdn.com/sd/punish/waf_block.html", "g.alicdn.com/dt/tracker/4.0.0/??tracker.Tracker.js", "'霸下通用 web 页面-验证码',  // 异常信息，推荐传入", "https://sbixi.alicdn.com/punish/punish:resource:template:", "腾讯T-Sec Web应用防火墙(WAF)", "提交的请求可能对网站造成威胁", "请求UUID为", "可能对网站造成安全威胁", "<div>您的请求ID是: ", "Attack request rejected", "background: url('https://errors.aliyun.com", "         Sorry, your request has been blocked as it may cause potential threats to the server's security"}

func filter(path string, resp *httpx.Response) bool {
    contentType := resp.RespHeader.Get("Content-Type")
    // 返回是个图片
    if util.Contains(contentType, "image/") {
        return true
    }
    
    if strings.HasSuffix(path, ".xml") {
        if !util.Contains(contentType, "xml") {
            
            return true
        }
    } else if strings.HasSuffix(path, ".json") {
        if !util.Contains(contentType, "json") {
            return true
        }
    }
    
    // 文件内容为空丢弃
    if resp.ContentLength == 0 {
        return true
    }
    
    title := util.GetTitle(resp.Body)
    
    if util.SliceInCaseFold(title, Page404Title) {
        return true
    }
    
    if util.SliceInCaseFold(resp.Body, Page404Content) {
        return true
    }
    
    if util.SliceInCaseFold(resp.Body, WafContent) {
        return true
    }
    
    return false
}
