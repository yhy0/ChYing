package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type GentooPlugin struct{}

func (p GentooPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "gentoo") {
            return true
        }
    }
    return false
}

func (p GentooPlugin) Name() string {
    return "Gentoo"
}
