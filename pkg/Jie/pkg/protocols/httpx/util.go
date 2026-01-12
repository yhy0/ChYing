package httpx

import (
	regexp "github.com/wasilibs/go-re2"
	"net"
	"net/url"
	"strings"
)

/**
   @author yhy
   @since 2025/6/7
   @desc //TODO
**/

func checkJSRedirect(htmlStr string) bool {
	redirectPatterns := []string{
		`window\.location\.href\s*=\s*['"][^'"]+['"]`,
		`window\.location\.assign\(['"][^'"]+['"]\)`,
		`window\.location\.replace\(['"][^'"]+['"]\)`,
		`window\.history\.(?:back|forward|go)\(`,
		`(?:setTimeout|setInterval)\([^,]+,\s*\d+\)`,
		`(?:onclick|onmouseover)\s*=\s*['"][^'"]+['"]`,
		`addEventListener\([^,]+,\s*function`,
		`(?ms)<a id="a-link"></a>\s*<script>\s*localStorage\.x5referer.*?document\.getElementById`,
	}

	for _, pattern := range redirectPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(htmlStr) {
			return true
		}
	}
	return false
}

// extractTitleFromHTML 从HTML响应中提取标题
func extractTitleFromHTML(body string) string {
	// 简单的标题提取，使用正则表达式
	titlePattern := `<title[^>]*>([^<]*)</title>`
	re := regexp.MustCompile(`(?i)` + titlePattern)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		// 限制标题长度
		if len(title) > 100 {
			title = title[:100] + "..."
		}
		return title
	}
	return ""
}

// extractIPFromTarget 从目标URL中提取IP地址
func extractIPFromTarget(target string) string {
	if u, err := url.Parse(target); err == nil {
		host := u.Hostname()
		// 简单检查是否是IP地址格式
		if net.ParseIP(host) != nil {
			return host
		}
		// 如果是域名，可以进行DNS解析，但这里为了性能考虑先返回空
		return ""
	}
	return ""
}

// extractPathFromTarget 从目标URL中提取路径
func extractPathFromTarget(target string) string {
	if u, err := url.Parse(target); err == nil {
		return u.RequestURI()
	}
	return "/"
}
