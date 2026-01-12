package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type PlayPlugin struct{}

func (p PlayPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "play! framework;") {
            return true
        }
    }

    return false
}

func (p PlayPlugin) Name() string {
    return "Play - Java Framework"
}
