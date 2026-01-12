package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type GrailsPlugin struct{}

func (p GrailsPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if _, ok := headers["X-Grails"]; ok {
        return true
    }

    if _, ok := headers["X-Grails-Cached"]; ok {
        return true
    }

    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "grails") {
            return true
        }
    }
    return false
}

func (p GrailsPlugin) Name() string {
    return "Grails - Java Framework"
}
