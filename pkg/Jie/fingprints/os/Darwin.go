package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type DarwinPlugin struct{}

func (p DarwinPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Darwin") {
            return true
        }
    }

    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "Darwin") {
            return true
        }
    }
    return false
}

func (p DarwinPlugin) Name() string {
    return "Darwin"
}
