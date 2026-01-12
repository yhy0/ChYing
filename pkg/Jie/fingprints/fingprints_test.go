package fingprints

import (
    "fmt"
    regexp "github.com/wasilibs/go-re2"
    "testing"
)

/**
  @author: yhy
  @since: 2023/10/12
  @desc: TODO
**/

func TestFingprints(t *testing.T) {
    headers := make(map[string][]string)
    
    headers["Generator"] = []string{"AsciiDoc 1.0.0;version:1.0.0"}
    headers["Set-Cookie"] = []string{"JSESSIONID="}
    headers["X-Powered-By"] = []string{"PHP/5.6.40-38+ubuntu20.04.1+deb.sury.org+1"}
    for _, p := range ProgramingPlugins {
        if p.Fingerprint("", headers) {
            t.Log(p.Name())
        }
    }
    
    re := regexp.MustCompile(`<\w[^>]*(="/_jcr_content/)[^>]*>`)
    if re.FindStringIndex("_jcr_content") != nil {
        fmt.Println("-==")
    }
}
