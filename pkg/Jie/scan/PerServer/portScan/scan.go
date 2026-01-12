package portScan

import (
    "context"
    "fmt"
    "github.com/Ullaakut/nmap/v2"
    "github.com/yhy0/ChYing/pkg/Jie/scan/PerServer/portScan/masscan"
    "github.com/yhy0/logging"
    "strings"
    "time"
)

/**
  @author: yhy
  @since: 2022/4/14
  @desc: //TODO
**/

func Scan(target, ip string) map[int]string {
    portService := make(map[int]string)
    
    logging.Logger.Println(portService)
    
    // 首先使用 masscan 快速探测 , 然后使用 nmap 进行验证和服务识别
    tmpPort := Masscan(target, ip)
    
    if len(tmpPort) > 0 {
        // nmap 扫描
        nmapRes := Nmap(ip, tmpPort)
        for _, host := range nmapRes {
            if host.Ports == nil || len(host.Ports) == 0 || len(host.Addresses) == 0 {
                continue
            }
            
            portMsg := fmt.Sprintf("\n\t\tHost %s - %q:\n\t\t", target, ip)
            
            for j, port := range host.Ports {
                if j == 0 {
                    portMsg += fmt.Sprintf("\t\t\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name)
                } else {
                    portMsg += fmt.Sprintf("\t\t\t\t\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name)
                }
            }
            
            logging.Logger.Infof("%s", portMsg)
            
            for _, port := range host.Ports {
                if port.State.State == "closed" {
                    continue
                }
                
                portService[int(port.ID)] = port.Service.Name
            }
        }
    }
    
    return portService
}

func Masscan(domain, ip string) []string {
    var ports []string
    m := masscan.New()
    m.SetSystemPath("masscan")
    m.SetRate("5000")
    
    args := []string{
        ip,
        "-p", "1-65535",
    }
    
    m.SetArgs(args...)
    err := m.Run()
    if err != nil {
        logging.Logger.Fatalln("masscan run err", err)
    }
    results, err := m.Parse()
    if err != nil {
        logging.Logger.Error(err)
    }
    
    // 如果某次masscan扫描的结果大于 40, 认为该ip做了策略，排除
    if len(results) > 40 {
        logging.Logger.Debugf("[%s - %s] 扫描出 %d 个端口, 排除.", domain, ip, len(results))
        return nil
    }
    
    for _, result := range results {
        ports = append(ports, result.Ports[0].Portid)
    }
    
    logging.Logger.Infof("The Masscan found %d ports of %s - %s", len(ports), domain, ip)
    return ports
}

func Nmap(ip string, ports []string) []nmap.Host {
    ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
    defer cancel()
    
    s, err := nmap.NewScanner(
        nmap.WithTargets(ip),
        nmap.WithSYNScan(),
        nmap.WithPorts(ports...),
        nmap.WithContext(ctx),
        nmap.WithSkipHostDiscovery(), // 加上 -Pn 就不去ping主机，因为有的主机防止ping,增加准确度
        nmap.WithDisabledDNSResolution(),
    )
    if err != nil {
        logging.Logger.Errorln("unable to create nmap scanner: %v", err)
        return nil
    }
    
    logging.Logger.Infof("Nmap => %+v", s.Args())
    result, warnings, err := s.Run()
    if len(warnings) > 0 {
        logging.Logger.Warningln(warnings)
        logging.Logger.Warning("You requested a scan type which requires root privileges, automatically switch scanning mode")
        if strings.Contains(warnings[0], "You requested a scan type which requires root privileges") {
            s, err = nmap.NewScanner(
                nmap.WithTargets(ip),
                nmap.WithPorts(ports...),
                nmap.WithContext(ctx),
                nmap.WithSkipHostDiscovery(), // 加上 -Pn 就不去ping主机，因为有的主机防止ping,增加准确度
                nmap.WithDisabledDNSResolution(),
            )
            result, warnings, err = s.Run()
        }
        
    }
    
    if err != nil {
        logging.Logger.Errorln("Unable to run nmap scan: %v  %v %v %v", err, warnings, ip, ports)
        return nil
    }
    
    return result.Hosts
}
