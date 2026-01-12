package programing

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type JavaPlugin struct{}

func (p JavaPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Set-Cookie"]; ok {
        if strings.Contains(strings.Join(v, ""), "JSESSIONID") {
            return true
        }
    }
    return false
}

func (p JavaPlugin) Name() string {
    return "Java"
}
