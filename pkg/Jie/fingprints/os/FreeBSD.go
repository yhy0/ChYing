package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type FreeBSDPlugin struct{}

func (p FreeBSDPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "FreeBSD") {
            return true
        }
    }

    return false
}

func (p FreeBSDPlugin) Name() string {
    return "FreeBSD"
}
