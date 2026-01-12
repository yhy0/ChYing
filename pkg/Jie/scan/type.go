package scan

import (
    "github.com/yhy0/ChYing/pkg/Jie/pkg/input"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
)

/**
   @author yhy
   @since 2023/10/13
   @desc //TODO
**/

// Addon 插件接口
type Addon interface {
    Scan(target string, path string, in *input.CrawlResult, client *httpx.Client) // 扫描, target\path 扫描目标单独传入，不从 in 中获取，这样就不用修改 in 中的 url 导致出现错误
    IsScanned(uniqueId string) bool                                               // 是否已经扫描过
    Name() string                                                                 // 插件名称
}
