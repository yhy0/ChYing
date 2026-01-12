package framework

import "strings"

/**
   @author yhy
   @since 2023/11/1
   @desc //TODO
**/

type BeegoPlugin struct{}

func (p BeegoPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if strings.Contains(body, "Powered by beego") {
        return true
    } else {
        if v, ok := headers["Set-Cookie"]; ok {
            if strings.Contains(strings.Join(v, ""), "beegosessionID=") {
                return true
            }
        }
    }

    return false
}

func (p BeegoPlugin) Name() string {
    return "Beego Web Framework - Go"
}
