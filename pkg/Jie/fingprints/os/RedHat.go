package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type RedHatPlugin struct{}

func (p RedHatPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Red Hat") {
            return true
        }
    }

    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "Red Hat") {
            return true
        }
    }
    return false
}

func (p RedHatPlugin) Name() string {
    return "RedHat"
}
