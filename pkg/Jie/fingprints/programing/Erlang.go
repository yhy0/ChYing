package programing

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type ErlangPlugin struct{}

func (p ErlangPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Erlang") {
            return true
        }
    }
    return false
}

func (p ErlangPlugin) Name() string {
    return "Erlang"
}
