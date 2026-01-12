package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type LaravelPlugin struct{}

func (p LaravelPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "laravel_session=") {
            return true
        }
    }

    return false
}

func (p LaravelPlugin) Name() string {
    return "Laravel - PHP Framework"
}
