package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type CodeIgniterPlugin struct{}

func (p CodeIgniterPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "ci_session=") {
            return true
        }
    }
    return false
}

func (p CodeIgniterPlugin) Name() string {
    return "CodeIgniter - PHP Framework"
}
