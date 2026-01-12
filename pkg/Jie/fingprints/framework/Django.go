package framework

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type DjangoPlugin struct{}

func (p DjangoPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "wgiserver/") || strings.Contains(value, "python/") || strings.Contains(value, "csrftoken=") {
            return true
        }
    }

    if strings.Contains(body, "<meta name=\"robots\" content=\"NONE,NOARCHIVE\"><title>Welcome to Django</title>") {
        return true
    }

    return false
}

func (p DjangoPlugin) Name() string {
    return "Django - Python Framework"
}
