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

type AspMvcPlugin struct{}

func (p AspMvcPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if _, ok := headers["X-AspNetMvc-Version"]; ok {
        return true
    }
    
    if _, ok := headers["X-AspNet-Version"]; ok {
        return true
    }
    
    re := regexp.MustCompile(`asp.net|anonymousID=|chkvalues=|__requestverificationtoken`)
    
    for _, v := range headers {
        value := strings.Join(v, "")
        if re.FindStringIndex(value) != nil {
            return true
        }
    }
    
    if strings.Contains(body, "Web Settings for Active Server Pages") {
        return true
    }
    
    if strings.Contains(body, "name=\"__VIEWSTATEENCRYPTED\" id=\"__VIEWSTATEENCRYPTED\"") {
        return true
    }
    
    return false
}

func (p AspMvcPlugin) Name() string {
    return "ASP.NET Framework"
}
