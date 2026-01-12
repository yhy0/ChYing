package framework

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type FlaskPlugin struct{}

func (p FlaskPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "flask") {
            return true
        }
    }

    return false
}

func (p FlaskPlugin) Name() string {
    return "Flask - Python Framework"
}
