package framework

import (
    regexp "github.com/wasilibs/go-re2"
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type SeagullPlugin struct{}

func (p SeagullPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if strings.Contains(body, "<meta name=\"generator\" content=\"Seagull Framework\" />") {
        return true
    }
    
    re := regexp.MustCompile(`Powered by <a href="http://seagullproject.org[/]*" title="Seagull framework homepage">Seagull PHP Framework</a>`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    
    re = regexp.MustCompile(`var SGL_JS_SESSID[\s]*=`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    
    return false
}

func (p SeagullPlugin) Name() string {
    return "Seagull - PHP Framework"
}
