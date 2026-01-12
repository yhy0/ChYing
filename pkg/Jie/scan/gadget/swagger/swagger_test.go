package swagger

import (
    "fmt"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "sync"
    "testing"
)

/**
  @author: yhy
  @since: 2023/1/4
  @desc: //TODO
**/

func TestSwagger(t *testing.T) {
    // 使用 sync.WaitGroup 防止 OutChannel 中的数据没有完全被消费，导致的数据漏掉问题
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        wg.Done()
        for v := range output.OutChannel {
            fmt.Println(v.PrintScreen())
        }
    }()
    
    Scan("", "")
    
    wg.Wait()
}
