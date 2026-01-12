package programing

import (
    regexp "github.com/wasilibs/go-re2"
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type PHPPlugin struct{}

func (p PHPPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "PHP/") {
            return true
        }
    }
    if v, ok := headers["Set-Cookie"]; ok {
        if strings.Contains(strings.Join(v, ""), "PHPSESSID") {
            return true
        }
    }
    
    if v, ok := headers["X-Powered-By"]; ok {
        if strings.Contains(strings.Join(v, ""), "PHP/") {
            return true
        }
    }
    
    re := regexp.MustCompile(`\.php(?:$|\?)`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    
    return false
}

func (p PHPPlugin) Name() string {
    return "PHP"
}
