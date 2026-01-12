package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type WindowsServerPlugin struct{}

func (p WindowsServerPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Win32") || strings.Contains(strings.Join(v, ""), "Win64") {
            return true
        }
    }
    return false
}

func (p WindowsServerPlugin) Name() string {
    return "Windows Server"
}
