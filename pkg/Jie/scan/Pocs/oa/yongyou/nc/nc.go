package nc

import (
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2022/9/21
  @desc: 用友 nc <=6.5   https://github.com/Ghost2097221/YongyouNC-Unserialize-Tools
**/

var paths = []string{
	"/servlet/~ic/nc.bs.framework.mx.monitor.MonitorServlet",
	"/servlet/~ic/nc.bs.framework.mx.MxServlet",
	"/servlet/~uapxbrl/uap.xbrl.persistenceImpl.XbrlPersistenceServlet",
	"/servlet/~uapss/com.yonyou.ante.servlet.FileReceiveServlet",
	"/servlet/~ic/nc.document.pub.fileSystem.servlet.DownloadServlet",
	"/servlet/~ic/nc.document.pub.fileSystem.servlet.UploadServlet",
	"/servlet/~ic/nc.document.pub.fileSystem.servlet.DeleteServlet",
	"/servlet/~ic/com.ufida.zior.console.ActionHandlerServlet",
	"/ServiceDispatcherServlet",
	"/servlet/~ic/bsh.servlet.BshServlet",
	"/servlet/~ic/ShowAlertFileServlet",
}

func Scan(u string, client *httpx.Client) bool {
	var resp *httpx.Response
	var err error
	for _, path := range paths {
		if path == "/ServiceDispatcherServlet" {
			resp, err = client.Request(u+path, "POST", "", nil, "Poc YongYou")
			if err != nil {
				continue
			}
		} else {
			resp, err = client.Request(u+path, "GET", "", nil, "Poc YongYou")
			if err != nil {
				continue
			}
		}

		// if resp.StatusCode == 200 || resp.StatusCode == 302 {
		if resp.StatusCode == 200 {
			logging.Logger.Infoln("接口存在", path)
			return true
		}
	}

	return false
}
