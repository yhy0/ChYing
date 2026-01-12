package os

import (
    regexp "github.com/wasilibs/go-re2"
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type DebianPlugin struct{}

func (p DebianPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Debian") {
            return true
        }
    }
    
    if v, ok := headers["X-Powered-By"]; ok {
        re := regexp.MustCompile(`(?:Debian|dotdeb|(sarge|etch|lenny|squeeze|wheezy|jessie))`)
        if re.FindStringIndex(strings.Join(v, "")) != nil {
            return true
        }
    }
    return false
}

func (p DebianPlugin) Name() string {
    return "Debian"
}
