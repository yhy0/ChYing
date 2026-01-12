package framework

import (
    "strings"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type YiiPlugin struct{}

func (p YiiPlugin) Fingerprint(body string, headers map[string][]string) bool {
    if strings.Contains(body, "<a href=\"http://www.yiiframework.com/\" rel=\"external\">Yii Framework</a>") {
        return true
    }

    if strings.Contains(body, ">Yii Framework</a>") {
        return true
    }

    return false
}

func (p YiiPlugin) Name() string {
    return "Yii - PHP Framework"
}
