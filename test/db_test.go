package test

import (
    "fmt"
    "github.com/iancoleman/orderedmap"
    "github.com/yhy0/ChYing/conf/file"
    "github.com/yhy0/logging"
    "testing"
)

/**
   @author yhy
   @since 2024/9/2
   @desc //TODO
**/

func TestNeo4j(t *testing.T) {
    logging.Logger = logging.New(true, file.ChyingDir, "ChYing", true)
    // db.Init("", "")
    
    o := orderedmap.New()
    
    o.Set("a", 1)
    
    // add some value with special characters
    o.Set("b", "\\.<>[]{}_-")
    
    o.Set("c", 3)
    for _, k := range o.Keys() {
        val, _ := o.Get(k)
        fmt.Println(k, val)
    }
    
    o.Set("a", 5)
    for _, k := range o.Keys() {
        val, _ := o.Get(k)
        fmt.Println(k, val)
    }
    
    // // 创建请求参数
    // request1 := map[string]string{
    //     "url":       "http://www.example.com/api/v1",
    //     "method":    "GET",
    //     "subdomain": "www.example.com",
    //     "domain":    "example.com",
    // }
    // params1 := map[string]string{
    //     "key1": "value1",
    //     "key2": "value2",
    //     "key3": "value3",
    // }
    //
    // request2 := map[string]string{
    //     "url":       "http://aaa.example.com/api/v2",
    //     "method":    "GET",
    //     "subdomain": "aaa.example.com",
    //     "domain":    "example.com",
    // }
    //
    // params2 := map[string]string{
    //     "key1": "value1",
    //     "key4": "value2",
    //     "key5": "value3",
    // }
    //
    // request3 := map[string]string{
    //     "url":       "http://aaa.example.com/api/v3",
    //     "method":    "GET",
    //     "subdomain": "aaa.example.com",
    //     "domain":    "example.com",
    // }
    //
    // params3 := map[string]string{
    //     "key6": "value1",
    //     "key7": "value2",
    // }
    //
    // request4 := map[string]string{
    //     "url":       "http://aaa.jd.com/api/v1",
    //     "method":    "GET",
    //     "subdomain": "aaa.jd.com",
    //     "domain":    "jd.com",
    // }
    //
    // params4 := map[string]string{
    //     "key8": "value1",
    // }
    //
    // db.InsertRRNode(request1, params1)
    // db.InsertRRNode(request2, params2)
    // db.InsertRRNode(request3, params3)
    // db.InsertRRNode(request4, params4)
}
