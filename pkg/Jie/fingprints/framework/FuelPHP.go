package framework

import "strings"

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

type FuelPHPPlugin struct{}

func (p FuelPHPPlugin) Fingerprint(body string, headers map[string][]string) bool {
    for _, v := range headers {
        value := strings.Join(v, "")
        if strings.Contains(value, "fuelcid=") {
            return true
        }
    }

    if strings.Contains(body, "Powered by <a href=\"http://fuelphp.com\">FuelPHP</a>") {
        return true
    }

    return false
}

func (p FuelPHPPlugin) Name() string {
    return "FuelPHP - PHP Framework"
}
