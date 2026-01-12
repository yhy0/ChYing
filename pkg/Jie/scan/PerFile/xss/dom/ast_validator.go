package dom

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/parser"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/logging"
)

// ASTValidator 是DOM XSS的第二层筛选器，使用AST和语法分析来验证预筛选的结果。
type ASTValidator struct {
	parser types.Parser
}

// NewASTValidator 创建一个新的AST验证器。
func NewASTValidator() *ASTValidator {
	return &ASTValidator{
		parser: parser.NewParser(),
	}
}

// ASTValidationResult 封装了AST验证阶段的结果。
type ASTValidationResult struct {
	IsStillSuspicious bool     // 经过AST验证后是否仍然可疑
	ValidatedSources  []string // 在合法上下文中找到的Source
	ValidatedSinks    []string // 在合法上下文中找到的Sink
}

// Validate 对预筛选的结果进行更深层次的AST验证。
// 它检查在预筛选阶段找到的Source和Sink是否真的存在于可执行的JavaScript上下文中，
// 而不是在注释或字符串字面量中。
func (v *ASTValidator) Validate(content string, prefilterResult *PrefilterResult) *ASTValidationResult {
	result := &ASTValidationResult{}

	if !prefilterResult.IsSuspicious {
		// 如果第一层就觉得没问题，第二层直接跳过
		return result
	}

	// 使用我们现有的解析器来获取AST
	// 我们构造一个临时的Response对象，因为解析器需要它
	dummyResp := &httpx.Response{Body: content}
	parsedResp, err := v.parser.Parse(dummyResp)
	if err != nil {
		logging.Logger.Warnf("DOM XSS AST验证阶段解析失败: %v", err)
		return result
	}

	doc := parsedResp.GoqueryDoc
	if doc == nil {
		return result
	}

	validatedSourcesMap := make(map[string]bool)
	validatedSinksMap := make(map[string]bool)

	// 在所有脚本块中验证 Sinks 和 Sources
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent := s.Text()
		if scriptContent == "" {
			return // 跳过空脚本块
		}

		// 验证 Sinks
		for _, sink := range prefilterResult.FoundSinks {
			if v.isIdentifierInScript(scriptContent, sink) {
				if !validatedSinksMap[sink] {
					result.ValidatedSinks = append(result.ValidatedSinks, sink)
					validatedSinksMap[sink] = true
				}
			}
		}

		// 验证 Sources
		for _, source := range prefilterResult.FoundSources {
			if v.isIdentifierInScript(scriptContent, source) {
				if !validatedSourcesMap[source] {
					result.ValidatedSources = append(result.ValidatedSources, source)
					validatedSourcesMap[source] = true
				}
			}
		}
	})

	// 最终决策：如果至少有一个经过验证的Source和一个Sink，则认为仍然可疑
	if len(result.ValidatedSources) > 0 && len(result.ValidatedSinks) > 0 {
		result.IsStillSuspicious = true
	}

	return result
}

// isIdentifierInScript 使用JS词法分析器，精确检查一个关键字是否存在于可执行代码中。
func (v *ASTValidator) isIdentifierInScript(scriptContent, keyword string) bool {
	l := js.NewLexer(parse.NewInputString(scriptContent))
	for {
		tt, text := l.Next()
		if tt == js.ErrorToken {
			break // 到达文件末尾或发生错误
		}

		// 我们寻找的是作为代码一部分的关键字（例如 `location.href` 中的 `location`）
		if strings.Contains(string(text), keyword) {
			// 关键判断：确保它不是字符串或注释的一部分
			if tt != js.StringToken && tt != js.CommentToken && tt != js.RegExpToken {
				return true
			}
		}
	}
	return false
}
