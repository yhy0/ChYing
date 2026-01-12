package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type UNIXPlugin struct{}

func (p UNIXPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "UNIX") {
            return true
        }
    }

    return false
}

func (p UNIXPlugin) Name() string {
    return "UNIX"
}
