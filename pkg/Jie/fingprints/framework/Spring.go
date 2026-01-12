package framework

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type SpringPlugin struct{}

func (p SpringPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "org.springframework.web.servlet.i18n.CookieLocaleResolver.LOCALE=") {
            return true
        }
    }

    return false
}

func (p SpringPlugin) Name() string {
    return "Spring Framework - Java Platform"
}
