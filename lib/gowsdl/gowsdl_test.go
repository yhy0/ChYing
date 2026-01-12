package gowsdl

import (
    "fmt"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "testing"
)

func TestGoWSDL(t *testing.T) {
    conf.GlobalConfig.Http.MaxQps = 5
    g, err := NewGoWSDL("http://127.0.0.1/dvwsuserservice?wsdl")
    
    if err != nil {
        t.Error(err)
    }
    
    fmt.Println(string(g.rawWSDL))
}
