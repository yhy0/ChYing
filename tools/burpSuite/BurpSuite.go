package burpSuite

import (
	"context"
	"fmt"
	"github.com/yhy0/ChYing/tools/burpSuite/mitmproxy/proxy"
	"github.com/yhy0/logging"
	"sync"
)

/**
  @author: yhy
  @since: 2023/4/24
  @desc: //TODO
**/

type SMap struct {
	sync.RWMutex
	Map map[int]*HTTPBody
}

func (l *SMap) WriteMap(key int, value *HTTPBody) {
	l.Lock()
	l.Map[key] = value
	l.Unlock()
}

func (l *SMap) ReadMap(key int) *HTTPBody {
	l.RLock()
	value, _ := l.Map[key]
	l.RUnlock()
	return value
}

// HttpHistory 接受 mitmproxy 代理信息
var HttpHistory chan HTTPHistory

// HTTPBodyMap 存储 mitmproxy 的响应信息, 为什么不直接放到HttpHistory，是为了防止太多的请求/响应数据加载到前端，这样做只有前端点击每行数据时才会加载对应的数据到前端显示
var HTTPBodyMap *SMap

// RepeaterBodyMap Repeater 中回退、前进使用 todo前端还未实现
var RepeaterBodyMap map[string]map[int]*HTTPBody

var IntruderMap map[string]*SMap

var Proxy *proxy.Proxy

var Intercept bool

func init() {
	HttpHistory = make(chan HTTPHistory, 1)

	HTTPBodyMap = &SMap{
		Map: make(map[int]*HTTPBody),
	}

	IntruderMap = make(map[string]*SMap)

	RepeaterBodyMap = make(map[string]map[int]*HTTPBody)
}

func Run(port string) {
	opts := &proxy.Options{
		Debug:             2,
		Addr:              fmt.Sprintf(":%s", port),
		StreamLargeBodies: 1024 * 1024 * 5,
		SslInsecure:       false,
		CaRootPath:        "",
	}

	var err error
	Proxy, err = proxy.NewProxy(opts)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// 这种不错，通过添加插件的形式，这样只要实现了接口,p.AddAddon(xxxx), 然后就会自动执行相应的操作
	// 添加一个日志记录插件
	Proxy.AddAddon(&proxy.LogAddon{})

	Proxy.AddAddon(&Burp{})

	logging.Logger.Errorln(Proxy.Start())
}

// Restart 更换监听端口
func Restart(port string) string {
	// 先关闭然后再启动
	err := Proxy.Shutdown(context.TODO())
	if err != nil {
		logging.Logger.Errorln(err)
		return err.Error()
	}
	go Run(port)
	return ""
}
