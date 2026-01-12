package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type SUSEPlugin struct{}

func (p SUSEPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "SUSE") {
            return true
        }
    }

    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "SUSE") {
            return true
        }
    }
    return false
}

func (p SUSEPlugin) Name() string {
    return "SunOS"
}
