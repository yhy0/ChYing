package burpSuite

import (
	"github.com/yhy0/ChYing/tools/burpSuite/mitmproxy/burp"
	"github.com/yhy0/ChYing/tools/burpSuite/mitmproxy/proxy"
	"github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2023/4/24
  @desc: //TODO
**/

func Run() {
	opts := &proxy.Options{
		Debug:             2,
		Addr:              ":9080",
		StreamLargeBodies: 1024 * 1024 * 5,
		SslInsecure:       false,
		CaRootPath:        "",
	}

	p, err := proxy.NewProxy(opts)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// 这种不错，通过添加插件的形式，这样只要实现了接口,p.AddAddon(xxxx), 然后就会自动执行相应的操作
	// 添加一个日志记录插件
	p.AddAddon(&proxy.LogAddon{})

	p.AddAddon(&burp.Burp{})

	logging.Logger.Fatal(p.Start())
}
