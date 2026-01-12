package parser

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/logging"
	"golang.org/x/net/html"
)

// defaultParser 是 types.Parser 接口的默认实现
type defaultParser struct{}

// NewParser 创建一个新的解析器实例
func NewParser() types.Parser {
	return &defaultParser{}
}

// Parse 从响应中解析出可供分析的结构
func (p *defaultParser) Parse(resp *httpx.Response) (types.ParsedResponse, error) {
	logging.Logger.Debugf("开始解析响应: %s", resp.RequestUrl)

	var parsed types.ParsedResponse
	parsed.Body = resp.Body

	// 1. 使用 golang.org/x/net/html 解析
	htmlNode, err := html.Parse(strings.NewReader(resp.Body))
	if err != nil {
		logging.Logger.Warnf("HTML标准解析失败: %v", err)
		// 即使标准解析失败，也继续尝试goquery，它可能更健壮
	}
	parsed.HTMLAstRoot = htmlNode

	// 2. 使用 github.com/PuerkitoBio/goquery 解析
	goqueryDoc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.Body))
	if err != nil {
		logging.Logger.Errorf("goquery解析失败: %v", err)
		return parsed, err // goquery失败通常意味着文档严重损坏，直接返回错误
	}
	parsed.GoqueryDoc = goqueryDoc

	// 3. 从goquery文档中提取所有脚本块
	var scripts []string
	goqueryDoc.Find("script").Each(func(i int, s *goquery.Selection) {
		// 优先获取外部脚本的src
		if src, exists := s.Attr("src"); exists {
			scripts = append(scripts, src)
		}
		// 获取内联脚本内容
		scripts = append(scripts, s.Text())
	})
	parsed.Scripts = scripts
	logging.Logger.Debugf("响应解析完成, 提取到 %d 个脚本块", len(scripts))

	return parsed, nil
}

// renderNode 将html.Node渲染回字符串，用于调试或特定场景
func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	err := html.Render(&buf, n)
	if err != nil {
		return ""
	}
	return buf.String()
}
