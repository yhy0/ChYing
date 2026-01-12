package programing

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type AsciiDocPlugin struct{}

func (p AsciiDocPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Generator"]; ok {
        if strings.Contains(strings.Join(v, ""), "AsciiDoc ") {
            return true
        }
    }
    return false
}

func (p AsciiDocPlugin) Name() string {
    return "AsciiDoc"
}
