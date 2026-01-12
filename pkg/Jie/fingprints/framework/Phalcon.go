package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type PhalconPlugin struct{}

func (p PhalconPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "phalcon-auth-") || strings.Contains(value, "phalconphp.com") || strings.Contains(value, "phalcon") {
            return true
        }
    }

    return false
}

func (p PhalconPlugin) Name() string {
    return "Phalcon - PHP Framework"
}
