package programing

import (
    regexp "github.com/wasilibs/go-re2"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type AspPlugin struct{}

func (p AspPlugin) Fingerprint(body string, headers map[string][]string) bool {
    re := regexp.MustCompile(`\.asp(x)?`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    
    return false
}

func (p AspPlugin) Name() string {
    return "ASP"
}
