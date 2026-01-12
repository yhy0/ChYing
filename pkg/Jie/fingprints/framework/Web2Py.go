package framework

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type Web2PyPlugin struct{}

func (p Web2PyPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "web2py") {
            return true
        }
    }

    if strings.Contains(body, "<div id=\"serendipityLeftSideBar\">") {
        return true
    }
    return false
}

func (p Web2PyPlugin) Name() string {
    return "Web2Py - Python Framework"
}
