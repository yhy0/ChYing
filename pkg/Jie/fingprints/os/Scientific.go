package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type ScientificPlugin struct{}

func (p ScientificPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Scientific Linux") {
            return true
        }
    }

    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "Scientific Linux") {
            return true
        }
    }
    return false
}

func (p ScientificPlugin) Name() string {
    return "Scientific"
}
