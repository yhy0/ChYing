package scan

import (
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/cmdinject"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/fastjson"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/jsonp"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/sql"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/sql/sqlmap"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/ssrf"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xss"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFile/xxe"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFolder/crlf"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFolder/log4j"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerFolder/traversal"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerServer"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerServer/portScan"
    "github.com/yhy0/ChYing/pkg/Jie/scan/bbscan"
    "github.com/yhy0/ChYing/pkg/Jie/scan/gadget/bypass403"
    "github.com/yhy0/ChYing/pkg/Jie/scan/gadget/collection"
)

/**
   @author yhy
   @since 2023/10/13
   @desc 这些基本漏洞插件化
**/

// PerFilePlugins 每个链接要测试的插件
var PerFilePlugins = make(map[string]Addon)

// PerFolderPlugins 每个目录要测试的插件
var PerFolderPlugins = make(map[string]Addon)

// PerServerPlugins 每个网站只测试一次的插件
var PerServerPlugins = make(map[string]Addon)

// 注册插件 , 每新增一个插件，这里都要注册一下
func init() {
    PerFilePlugins["xss"] = &xss.Plugin{}
    PerFilePlugins["sql"] = &sql.Plugin{}
    PerFilePlugins["sqlmapApi"] = &sqlmap.Plugin{}
    PerFilePlugins["ssrf"] = &ssrf.Plugin{}
    PerFilePlugins["jsonp"] = &jsonp.Plugin{}
    PerFilePlugins["cmd"] = &cmdinject.Plugin{}
    PerFilePlugins["xxe"] = &xxe.Plugin{}
    PerFilePlugins["fastjson"] = &fastjson.Plugin{}
    PerFilePlugins["bypass403"] = &bypass403.Plugin{}
    
    PerFilePlugins["SensitiveParameters"] = &collection.Plugin{} // 这个不受开关控制
    
    PerFolderPlugins["crlf"] = &crlf.Plugin{}
    PerFolderPlugins["iis"] = &crlf.Plugin{}
    PerFolderPlugins["nginx-alias-traversal"] = &traversal.Plugin{}
    PerFolderPlugins["log4j"] = &log4j.Plugin{}
    PerFolderPlugins["bbscan"] = &bbscan.Plugin{} // 扫描规则路径不是 root 的需要扫描
    
    PerServerPlugins["bbscan"] = &bbscan.Plugin{}
    PerServerPlugins["portScan"] = &portScan.Plugin{}
    // PerServerPlugins["nuclei"] = &PerServer.NucleiPlugin{}
    PerServerPlugins["archive"] = &PerServer.ArchivePlugin{}
}
