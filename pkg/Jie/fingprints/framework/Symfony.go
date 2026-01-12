package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type SymfonyPlugin struct{}

func (p SymfonyPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if strings.Contains(body, "\"powered by symfony\"") {
        return true
    }

    if strings.Contains(body, "Powered by <a href=\"http://www.symfony-project.org/\">") {
        return true
    }

    return false
}

func (p SymfonyPlugin) Name() string {
    return "Symfony - PHP Framework"
}
