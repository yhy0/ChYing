package struts2

import (
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-001"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-005"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-007"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-008"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-009"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-012"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-013"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-015"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-016"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-045"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-046"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-048"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-053"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/s2-057"
    "github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/java/struts2/utils"
    "github.com/yhy0/logging"
    "log"
)

/**
   @author yhy
   @since 2023/8/19
   @desc //TODO
**/

func S2(options conf.Options, client *httpx.Client) {
    for _, target := range conf.GlobalConfig.Options.Targets {
        if options.S2.Mode == "scan" {
            switch options.S2.Name {
            case "s2-001":
                if options.S2.Body != "" {
                    s2_001.Check(target, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s001须指定POST数据包内容，并用<fuckit>标记出测试点，如: --options.S2.Body \"user=a&pass=fuckit\"")
                }
            case "s2-005":
                s2_005.Check(target)
            case "s2-007":
                if options.S2.Body != "" {
                    s2_007.Check(target, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s007需指定POST数据包内容，并用<fuckit>标记出测试点，如: --options.S2.Body \"user=a&pass=fuckit\"")
                }
            case "s2-008":
                s2_008.Check(target)
            case "s2-009":
                if options.S2.Body != "" {
                    s2_009.Check(target, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s009需指定要测试的GET参数，如: --options.S2.Body=\"name\"")
                }
            case "s2-012":
                if options.S2.Body != "" {
                    s2_012.Check(target, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s012需手动指定POST数据包内容，并用<fuckit>标记出测试点，如: --options.S2.Body \"user=a&pass=fuckit\"")
                }
            case "s2-013":
                s2_013.Check(target)
            case "s2-015":
                s2_015.Check(target)
            case "s2-016":
                s2_016.Check(target)
            case "s2-045":
                s2_045.Check(target)
            case "s2-046":
                s2_046.Check(target, client)
            case "s2-048":
                if options.S2.Body != "" {
                    s2_048.Check(target, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s048需手动指定POST数据包内容，并用<fuckit>标记出测试点，如: --options.S2.Body \"user=a&pass=fuckit\"")
                }
            case "s2-053":
                if options.S2.Body != "" {
                    s2_053.Check(target, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s053需手动指定POST数据包内容，并用<fuckit>标记出测试点，如: --options.S2.Body \"user=a&pass=fuckit\"")
                }
            case "s2-057":
                s2_057.Check(target)
            case "all":
                logging.Logger.Infoln("未指定漏洞编号,默认全检测")
                s2_001.Check(target, options.S2.Body)
                s2_005.Check(target)
                s2_007.Check(target, options.S2.Body)
                s2_008.Check(target)
                s2_009.Check(target, options.S2.Body)
                s2_012.Check(target, options.S2.Body)
                s2_013.Check(target)
                s2_015.Check(target)
                s2_016.Check(target)
                s2_045.Check(target)
                s2_046.Check(target, client)
                s2_048.Check(target, options.S2.Body)
                s2_053.Check(target, options.S2.Body)
                s2_057.Check(target)
            default:
                logging.Logger.Infoln("漏洞编号设置错误，目前支持检测：")
                for _, vnn := range utils.Vnlist {
                    logging.Logger.Infoln(vnn)
                }
            }
        } else if options.S2.Mode == "exec" && options.S2.CMD != "" {
            switch options.S2.Name {
            case "s2-001":
                if options.S2.Body != "" {
                    s2_001.ExecCommand(target, options.S2.CMD, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s001需手动指定POST数据包内容，并用fuckit标记出测试点，如: --options.S2.Body=\"user=a&pass=fuckit\"")
                }
            case "s2-005":
                s2_005.ExecCommand(target, options.S2.CMD)
            case "s2-007":
                if options.S2.Body != "" {
                    s2_007.ExecCommand(target, options.S2.CMD, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s007需手动指定POST数据包内容，并用fuckit标记出测试点，如: --options.S2.Body=\"user=a&pass=fuckit\"")
                }
            case "s2-008":
                s2_008.ExecCommand(target, options.S2.CMD)
            case "s2-009":
                s2_009.ExecCommand(target, options.S2.CMD, options.S2.Body)
            case "s2-012":
                if options.S2.Body != "" {
                    s2_012.ExecCommand(target, options.S2.CMD, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s012需手动指定POST数据包内容，并用<fuckit>标记出测试点，如: --options.S2.Body=\"user=a&pass=fuckit\"")
                }
            case "s2-013":
                s2_013.ExecCommand(target, options.S2.CMD)
            case "s2-015":
                s2_015.ExecCommand(target, options.S2.CMD)
            case "s2-016":
                s2_016.ExecCommand(target, options.S2.CMD)
            case "s2-045":
                s2_045.ExecCommand(target, options.S2.CMD)
            case "s2-046":
                s2_046.ExecCommand(target, options.S2.CMD, client)
            case "s2-047":
                if options.S2.Body != "" {
                    s2_048.ExecCommand(target, options.S2.CMD, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s048需手动指定POST数据包内容，并用<fuckit>标记出测试点，如: --options.S2.Body=\"user=a&pass=fuckit\"")
                }
            case "s2-053":
                if options.S2.Body != "" {
                    s2_053.ExecCommand(target, options.S2.CMD, options.S2.Body)
                } else {
                    logging.Logger.Infoln("s053需手动指定POST数据包内容，并用<fuckit>标记出测试点，如: --options.S2.Body=\"user=a&pass=fuckit\"")
                }
            case "s2-057":
                s2_057.ExecCommand(target, options.S2.CMD)
            case "all":
                log.Fatalf("命令执行模式必须指定漏洞编号")
            default:
                log.Fatalf("命令执行模式必须指定漏洞编号")
            }
        } else {
            logging.Logger.Infoln("参数错误")
        }
    }
    
}
