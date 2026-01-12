package programing

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type LuaPlugin struct{}

func (p LuaPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "Lua") {
            return true
        }
    }
    return false
}

func (p LuaPlugin) Name() string {
    return "Lua"
}
