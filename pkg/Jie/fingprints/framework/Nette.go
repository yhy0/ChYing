package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type NettePlugin struct{}

func (p NettePlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "Nette") || strings.Contains(value, "nette-browser=") {
            return true
        }
    }

    return false
}

func (p NettePlugin) Name() string {
    return "Nette - PHP Framework"
}
