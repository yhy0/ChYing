package fingprints

import (
    "github.com/yhy0/ChYing/pkg/Jie/conf"
)

/**
  @author: yhy
  @since: 2023/10/12
  @desc: 参考 https://github.com/w-digital-scanner/w13scan/tree/master/W13SCAN/fingprints
**/

func Identify(body []byte, headers map[string][]string) []string {
    var fingerprints []string
    // web app 指纹检测
    webapps := conf.Wappalyzer.Fingerprint(headers, body)
    
    for webapp := range webapps {
        fingerprints = append(fingerprints, webapp)
    }
    
    // web 框架指纹检测
    for _, p := range FrameworkPlugins {
        if p.Fingerprint(string(body), headers) {
            fingerprints = append(fingerprints, p.Name())
        }
    }
    
    // 语言指纹检测
    for _, p := range ProgramingPlugins {
        if p.Fingerprint(string(body), headers) {
            fingerprints = append(fingerprints, p.Name())
        }
    }
    
    // 系统指纹检测
    for _, p := range OsPlugins {
        if p.Fingerprint(string(body), headers) {
            fingerprints = append(fingerprints, p.Name())
        }
    }
    
    return fingerprints
}
