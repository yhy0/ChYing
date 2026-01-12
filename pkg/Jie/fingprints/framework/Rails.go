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

type RailsPlugin struct{}

func (p RailsPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if _, ok := headers["X-Rails"]; ok {
        return true
    }
    
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "phusion passenger") || strings.Contains(value, "rails") {
            return true
        }
    }
    
    re := regexp.MustCompile(`<meta content="authenticity_token" name="csrf-param"\s?/>\s?<meta content="[^"]{44}" name="csrf-token"\s?/>`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    
    re = regexp.MustCompile(`<link[^>]*href="[^"]*/assets/application-?\w{32}?\.css"`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    
    re = regexp.MustCompile(`<script[^>]*/assets/application-?\w{32}?\.js"`)
    if re.FindStringIndex(body) != nil {
        return true
    }
    
    return false
}

func (p RailsPlugin) Name() string {
    return "Ruby on Rails - Ruby Framework"
}
