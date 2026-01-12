package wsdl

import (
    "strings"
)

/**
   @author yhy
   @since 2024/12/23
   @desc //TODO
**/

func Payload(_type string) string {
    _type = strings.ToLower(_type)
    // fmt.Println("_type:", _type)
    var res = "test"
    if strings.Contains(_type, "string") || strings.Contains(_type, "token") {
        res = "test"
    } else if strings.Contains(_type, "long") || strings.Contains(_type, "short") || strings.Contains(_type, "int") || strings.Contains(_type, "integer") || strings.Contains(_type, "byte") || strings.Contains(_type, "float") || strings.Contains(_type, "double") || strings.Contains(_type, "decimal") {
        res = "11"
    } else if strings.Contains(_type, "boolean") {
        res = "false"
    }
    
    return res
}
