package framework

import (
    regexp "github.com/wasilibs/go-re2"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type ApacheJackrabbitPlugin struct{}

func (p ApacheJackrabbitPlugin) Fingerprint(body string, headers map[string][]string) bool {
    re := regexp.MustCompile(`<\w[^>]*(="/_jcr_content/)[^>]*>`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    return false
}

func (p ApacheJackrabbitPlugin) Name() string {
    return "Apache Jackrabbit/Adobe CRX repository"
}
