package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type KarrigellPlugin struct{}

func (p KarrigellPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "karrigell") {
            return true
        }
    }

    return false
}

func (p KarrigellPlugin) Name() string {
    return "Karrigell - Python Framework"
}
