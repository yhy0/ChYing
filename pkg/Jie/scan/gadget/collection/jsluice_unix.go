//go:build !windows

package collection

import (
	"github.com/yhy0/ChYing/lib/jsluice"
	"github.com/yhy0/logging"
)

// analyzeJsluice uses jsluice to extract URLs from JavaScript content
func analyzeJsluice(target, body string) []string {
	var res []string
	analyzer := jsluice.NewAnalyzer([]byte(body))
	for _, u := range analyzer.GetURLs() {
		logging.Logger.Debugln("[jsluice]", target, u.URL)
		res = append(res, u.URL)
	}
	return res
}
