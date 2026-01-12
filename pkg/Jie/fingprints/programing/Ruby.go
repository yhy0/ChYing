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

type RubyPlugin struct{}

func (p RubyPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if v, ok := headers["Server"]; ok {
        re := regexp.MustCompile(`(?:Mongrel|WEBrick|Ruby)`)
        if re.FindStringIndex(strings.Join(v, "")) != nil {
            return true
        }
    }
    return false
}

func (p RubyPlugin) Name() string {
    return "Ruby"
}
