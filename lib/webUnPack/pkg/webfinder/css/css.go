package css

import (
    "regexp"
    "strings"
)

// UriFromCss https://www.w3.org/TR/css-syntax-3/
func UriFromCss(csstext string) []string {
    var results []string
    var cssRex = `(?m)url\(['"]?(.+?)['"]?\)`
    
    csstext = strings.ReplaceAll(csstext, " ", "")
    csstext = strings.ReplaceAll(csstext, "\r", "")
    csstext = strings.ReplaceAll(csstext, "\n", "")
    csstext = strings.ReplaceAll(csstext, "\t", "")
    // fmt.Println(csstext)
    
    re := regexp.MustCompile(cssRex)
    matches := re.FindAllStringSubmatch(csstext, -1)
    for _, match := range matches {
        // fmt.Println(match[1])
        if strings.Contains(match[1], "data:") {
            continue
        }
        results = append(results, match[1])
    }
    return results
}
