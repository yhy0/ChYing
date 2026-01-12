package api

import (
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/output"
	"github.com/yhy0/ChYing/lib/webUnPack/pkg/utils"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/logging"
)

func Enumerate(target string, uri output.Result) {
	if uri.Value != "/" && uri.Value != "" {
		target = utils.Standard(target, uri.Value)
		res, err := httpx.Get(target, "WebUnPack")
		if err != nil {
			return
		}
		logging.Logger.Debugf("[%d] %s", res.StatusCode, target)
		if res.StatusCode >= 300 {
			return
		}
		logging.Logger.Infof("[%d] %s source:%s", res.StatusCode, target, uri.Source)
		// db.AddVul(db.Vulnerability{
		//	Target:    target,
		//	Plugin:    "WebUnpack",
		//	Response:  res.Body,
		//	Reference: uri.Source,
		// })
	}
}
