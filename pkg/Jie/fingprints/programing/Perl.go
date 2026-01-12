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

type PerlPlugin struct{}

func (p PerlPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        if strings.Contains(strings.Join(v, ""), "Perl") {
            return true
        }
    }
    
    re := regexp.MustCompile(`\.pl(?:$|\?)`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    
    return false
}

func (p PerlPlugin) Name() string {
    return "Perl"
}
