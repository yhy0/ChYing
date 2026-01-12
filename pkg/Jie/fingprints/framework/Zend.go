package framework

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type ZendPlugin struct{}

func (p ZendPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "zend") {
            return true
        }
    }

    if strings.Contains(body, "<a href=\"http://www.yiiframework.com/\" rel=\"external\">Yii Framework</a>") {
        return true
    }

    if strings.Contains(body, "<meta name=\"generator\" content=\"Zend.com CMS ") {
        return true
    }

    if strings.Contains(body, "\"Powered by Zend Framework\"") {
        return true
    }

    if strings.Contains(body, "alt=\"Powered by Zend Framework!\" />") {
        return true
    }

    return false
}

func (p ZendPlugin) Name() string {
    return "Yii - PHP Framework"
}
