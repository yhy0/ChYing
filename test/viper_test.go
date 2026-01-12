package test

import (
    "bytes"
    "fmt"
    "github.com/spf13/viper"
    JieConf "github.com/yhy0/ChYing/pkg/Jie/conf"
    "testing"
)

/**
   @author yhy
   @since 2024/9/24
   @desc //TODO
**/

func TestViper(t *testing.T) {
    viper.SetConfigType("yaml")
    var yamlExample = `
Hacker: true
name: steve
hobbies:
- skateboarding
- snowboarding
- go
clothing:
  jacket: leather
  trousers: denim
age: 35
eyes : brown
beard: true
`
    err := viper.ReadConfig(bytes.NewBufferString(yamlExample))
    if err != nil {
        t.Log(err)
        return
    }
    fmt.Println(viper.Get("name"))
    err = viper.Unmarshal(&JieConf.GlobalConfig)
    
    if err != nil {
        t.Log(err)
    }
    fmt.Println(JieConf.GlobalConfig)
}
