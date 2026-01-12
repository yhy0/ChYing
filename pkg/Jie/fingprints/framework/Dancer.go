package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type DancerPlugin struct{}

func (p DancerPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "Dancer") || strings.Contains(value, "dancer.session=") {
            return true
        }
    }
    return false
}

func (p DancerPlugin) Name() string {
    return "Dancer - Perl Framework"
}
