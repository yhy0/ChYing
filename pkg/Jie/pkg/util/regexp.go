package util

import (
    regexp "github.com/wasilibs/go-re2"
    "strings"
)

/**
  @author: yhy
  @since: 2023/2/10
  @desc: //TODO
**/

func MatchAnyOfRegexp(regexps []string, match string) (bool, string) {
    for _, value := range regexps {
        regex := regexp.MustCompile(value)
        if regex.MatchString(match) {
            return true, value
        }
    }
    
    return false, ""
}

func RegexpStr(patterns []string, str string) bool {
    for _, pattern := range patterns {
        if strings.TrimSpace(pattern) == "" {
            continue
        }
        match, err := regexp.MatchString(pattern, str)
        if err != nil {
            continue
        }
        if match {
            return true
        }
    }
    
    return false
}
