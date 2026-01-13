//go:build windows

package utils

// Jsluice is a stub for Windows platform where tree-sitter-javascript is not available.
// It returns empty results on Windows.
func Jsluice(body string, apis []string) ([]string, []string) {
	return []string{}, []string{}
}
