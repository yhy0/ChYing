package dom

import (
	"strings"

	"github.com/yhy0/logging"
)

// 定义已知的DOM XSS Sources 和 Sinks
// 来源：OWASP, PortSwigger, 以及社区研究
var domSources = []string{
	"location",
	"document.URL",
	"document.documentURI",
	"document.URLUnencoded",
	"document.baseURI",
	"location.href",
	"location.search",
	"location.hash",
	"location.pathname",
	"document.cookie",
	"document.referrer",
	"window.name",
	"history.pushState",
	"history.replaceState",
	"localStorage",
	"sessionStorage",
	"postMessage",
	"navigator.sendBeacon",
}

var domSinks = []string{
	// HTML Sinks
	"innerHTML",
	"outerHTML",
	"insertAdjacentHTML",
	"onevent", // 通用事件处理器

	// Script Execution Sinks
	"eval",
	"setTimeout",
	"setInterval",
	"setImmediate",
	"Function",
	"execScript",

	// URL Sinks
	"location.href",
	"location.replace",
	"location.assign",
	"open",
	"document.write",
	"document.writeln",
}

// PrefilterResult 静态预筛选的结果
type PrefilterResult struct {
	IsSuspicious          bool     // 是否可疑
	FoundSources          []string // 发现的Source
	FoundSinks            []string // 发现的Sink
	HasModernFramework    bool     // 是否包含现代框架 (React, Vue, Angular等)
	HasES6Features        bool     // 是否包含ES6+特性
	HasSPAFeatures        bool     // 是否为单页应用
	HasBuildToolSignature bool     // 是否有构建工具特征
	JSComplexityScore     int      // JS复杂度分数
	DetectedFrameworks    []string // 检测到的具体框架
}

// Prefilter 是一个轻量级的静态分析器，用于初步筛选潜在的DOM XSS目标
type Prefilter struct {
	// 现代框架检测模式
	frameworkPatterns []string
	es6Patterns       []string
	spaPatterns       []string
	buildToolPatterns []string
}

// NewPrefilter 创建一个新的预筛选器
func NewPrefilter() *Prefilter {
	p := &Prefilter{
		// 现代框架特征（借鉴 parser_selector.go 和 modern_analyzer.go 的精华）
		frameworkPatterns: []string{
			"react", "vue", "angular", "svelte", "next.js", "nuxt", "gatsby",
			"ng-app", "data-reactroot", "v-model", "v-if", "v-for",
			"ng-", "*ng", "jsx", "typescript", "_app_", "__next",
			"jquery", "backbone", "ember", "knockout",
		},
		// ES6+ 特性检测
		es6Patterns: []string{
			"const ", "let ", "=>", "`${", "async ", "await ",
			"class ", "import ", "export ", "...", "function*",
			"symbol", "proxy", "promise", "map", "set", "weakmap",
		},
		// SPA 特征检测
		spaPatterns: []string{
			"history.pushstate", "history.replacestate", "router",
			"single.page", "__app_data__", "app.js", "spa",
			"route:", "$router", "browserouter", "hashrouter",
		},
		// 构建工具特征
		buildToolPatterns: []string{
			"webpack", "vite", "rollup", "parcel", "babel",
			"__webpack_require__", "system.import", "esm",
			"chunk", "vendor", "manifest", ".min.js",
		},
	}
	return p
}

// Analyze 对给定的HTML/JS内容进行静态分析
// 它通过搜索已知的sources和sinks来判断页面是否值得进行更昂贵的动态分析
func (p *Prefilter) Analyze(content string) PrefilterResult {
	result := PrefilterResult{
		IsSuspicious: false,
		FoundSources: []string{},
		FoundSinks:   []string{},
	}

	foundSourcesMap := make(map[string]bool)
	foundSinksMap := make(map[string]bool)

	// 一次性转为小写以优化搜索
	contentLower := strings.ToLower(content)

	// 使用优化的批量搜索算法来提升性能
	p.performBatchSearch(contentLower, domSources, foundSourcesMap, &result.FoundSources)
	p.performBatchSearch(contentLower, domSinks, foundSinksMap, &result.FoundSinks)

	// 3. 现代技术特征检测 - 增强版检测逻辑
	result.HasModernFramework, result.DetectedFrameworks = p.detectFrameworksWithDetails(contentLower)
	result.HasES6Features = p.hasAnyPattern(contentLower, p.es6Patterns)
	result.HasSPAFeatures = p.hasAnyPattern(contentLower, p.spaPatterns)
	result.HasBuildToolSignature = p.hasAnyPattern(contentLower, p.buildToolPatterns)

	// 4. 计算增强的JS复杂度分数
	result.JSComplexityScore = p.calculateEnhancedComplexityScore(content, contentLower)

	// 5. 智能化可疑判断 - 考虑更多现代技术因素
	result.IsSuspicious = p.evaluateSuspiciousness(result)

	// 6. 输出检测摘要（仅在检测到现代技术时）
	if result.HasModernFramework || result.HasES6Features || result.HasSPAFeatures {
		p.logDetectionSummary(result)
	}

	return result
}

// performBatchSearch 执行批量搜索以提高性能
func (p *Prefilter) performBatchSearch(content string, patterns []string, foundMap map[string]bool, results *[]string) {
	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			if !foundMap[pattern] {
				*results = append(*results, pattern)
				foundMap[pattern] = true
			}
		}
	}
}

// hasAnyPattern 检查内容中是否包含任一模式
func (p *Prefilter) hasAnyPattern(content string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// detectFrameworksWithDetails 检测框架并返回详细信息
func (p *Prefilter) detectFrameworksWithDetails(content string) (bool, []string) {
	var detectedFrameworks []string

	// 检测主要框架
	frameworks := map[string][]string{
		"React":   {"react", "data-reactroot", "jsx", "_app_"},
		"Vue":     {"vue", "v-model", "v-if", "v-for"},
		"Angular": {"angular", "ng-app", "ng-", "*ng"},
		"Svelte":  {"svelte"},
		"Next.js": {"next.js", "__next"},
		"Nuxt":    {"nuxt"},
		"Gatsby":  {"gatsby"},
		"jQuery":  {"jquery", "$.(", "$(\"", "$('"},
	}

	for frameworkName, patterns := range frameworks {
		for _, pattern := range patterns {
			if strings.Contains(content, pattern) {
				detectedFrameworks = append(detectedFrameworks, frameworkName)
				break // 避免重复添加同一框架
			}
		}
	}

	return len(detectedFrameworks) > 0, detectedFrameworks
}

// calculateEnhancedComplexityScore 计算增强的复杂度分数
func (p *Prefilter) calculateEnhancedComplexityScore(content, contentLower string) int {
	score := 0

	// 基础脚本计数
	scriptCount := strings.Count(contentLower, "<script")
	score += scriptCount * 10

	// 行数复杂度
	lineBreaks := strings.Count(content, "\n")
	score += lineBreaks / 100

	// ES6+ 特性加分
	es6Count := 0
	for _, pattern := range p.es6Patterns {
		if strings.Contains(contentLower, pattern) {
			es6Count++
		}
	}
	score += es6Count * 5

	// SPA 特性加分
	spaCount := 0
	for _, pattern := range p.spaPatterns {
		if strings.Contains(contentLower, pattern) {
			spaCount++
		}
	}
	score += spaCount * 8

	// 构建工具特征加分
	buildToolCount := 0
	for _, pattern := range p.buildToolPatterns {
		if strings.Contains(contentLower, pattern) {
			buildToolCount++
		}
	}
	score += buildToolCount * 6

	return score
}

// evaluateSuspiciousness 智能化可疑性评估
func (p *Prefilter) evaluateSuspiciousness(result PrefilterResult) bool {
	// 传统判断：同时有 Source 和 Sink
	if len(result.FoundSources) > 0 && len(result.FoundSinks) > 0 {
		return true
	}

	// 现代框架应用的特殊判断
	if result.HasModernFramework && len(result.FoundSources) > 0 {
		return true
	}

	// SPA 应用的特殊判断
	if result.HasSPAFeatures && len(result.FoundSources) > 0 {
		return true
	}

	// 高复杂度且有 ES6 特性的应用
	if result.JSComplexityScore > 50 && result.HasES6Features && len(result.FoundSources) > 0 {
		return true
	}

	// 有构建工具特征且复杂度较高
	if result.HasBuildToolSignature && result.JSComplexityScore > 30 && len(result.FoundSources) > 0 {
		return true
	}

	return false
}

// logDetectionSummary 输出检测结果摘要
func (p *Prefilter) logDetectionSummary(result PrefilterResult) {
	if len(result.DetectedFrameworks) > 0 {
		logging.Logger.Debugf("【DOM预筛选】检测到现代框架: %v", result.DetectedFrameworks)
	}

	var features []string
	if result.HasES6Features {
		features = append(features, "ES6+")
	}
	if result.HasSPAFeatures {
		features = append(features, "SPA")
	}
	if result.HasBuildToolSignature {
		features = append(features, "构建工具")
	}

	if len(features) > 0 {
		logging.Logger.Debugf("【DOM预筛选】现代特性: %v, 复杂度: %d", features, result.JSComplexityScore)
	}
}
