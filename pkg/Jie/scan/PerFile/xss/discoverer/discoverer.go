package discoverer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss/types"
	"github.com/yhy0/logging"
)

// formDiscoverer implements the AttackSurfaceDiscoverer interface to find form inputs.
type formDiscoverer struct{}

// NewFormDiscoverer creates a new form discoverer.
func NewFormDiscoverer() types.AttackSurfaceDiscoverer {
	return &formDiscoverer{}
}

// Discover finds all input, textarea, and select fields within forms in an HTML response.
func (d *formDiscoverer) Discover(resp *httpx.Response) []httpx.Param {
	var discoveredParams []httpx.Param

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.Body))
	if err != nil {
		logging.Logger.Warnf("在攻击面发现中解析HTML失败: %v", err)
		return discoveredParams
	}

	foundParams := make(map[string]struct{})

	// Find inputs, textareas, and selects inside forms
	doc.Find("form").Each(func(i int, form *goquery.Selection) {
		form.Find("input, textarea, select").Each(func(j int, input *goquery.Selection) {
			name, exists := input.Attr("name")
			if !exists || name == "" {
				return // Skip inputs without a name
			}

			// Avoid adding duplicate parameter names
			if _, ok := foundParams[name]; ok {
				return
			}

			// Create a new parameter based on httpx.Param structure
			param := httpx.Param{
				Name:  name,
				Value: "ChYingRefactoredTest", // Use a default test value
				Index: j,                      // Set index for ordering
			}
			discoveredParams = append(discoveredParams, param)
			foundParams[name] = struct{}{}
		})
	})

	if len(discoveredParams) > 0 {
		logging.Logger.Infof("从响应HTML中发现了 %d 个新的表单参数", len(discoveredParams))
	}

	return discoveredParams
}
