//go:build windows

package collection

// analyzeJsluice is a stub for Windows platform where tree-sitter-javascript is not available.
// It returns empty results on Windows.
func analyzeJsluice(target, body string) []string {
	return []string{}
}
