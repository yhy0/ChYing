package os

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type FedoraPlugin struct{}

func (p FedoraPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Fedora") {
            return true
        }
    }
    return false
}

func (p FedoraPlugin) Name() string {
    return "Fedora"
}
