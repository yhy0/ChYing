package framework

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type CakePHPPlugin struct{}

func (p CakePHPPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "CAKEPHP=") {
            return true
        }
    }

    return false
}

func (p CakePHPPlugin) Name() string {
    return "CakePHP - PHP Framework"
}
