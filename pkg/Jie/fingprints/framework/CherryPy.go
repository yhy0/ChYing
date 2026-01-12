package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type CherryPyPlugin struct{}

func (p CherryPyPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "CherryPy") {
            return true
        }
    }
    return false
}

func (p CherryPyPlugin) Name() string {
    return "CherryPy - Python Framework"
}
