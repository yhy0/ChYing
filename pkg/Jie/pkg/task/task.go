package task

import (
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/panjf2000/ants/v2"
	"github.com/sasha-s/go-deadlock"
	"github.com/thoas/go-funk"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/conf"
	"github.com/yhy0/ChYing/pkg/Jie/fingprints"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/input"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/ChYing/pkg/Jie/scan"
	"github.com/yhy0/ChYing/pkg/Jie/scan/Pocs/pocs_go"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/bypass403"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/collection"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/jwt"
	"github.com/yhy0/ChYing/pkg/Jie/scan/gadget/sensitive"
	"github.com/yhy0/logging"
	"github.com/yhy0/sizedwaitgroup"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/**
  @author: yhy
  @since: 2023/1/5
  @desc:  ~~后期看看有没有必要设计成插件式的，自我感觉没必要，还不如这样写，逻辑简单易懂~~
        最终还是插件式比较好 ，哈哈
        todo 漏洞检测逻辑有待优化, 每个插件扫描到漏洞后，需要及时退出，不再进行后续扫描, 插件内部应该设置一个通知，扫描到漏洞即停止
**/

type Task struct {
	Fingerprints []string             // 这个只有主动会使用，被动只会新建一个 task，所以不会用到
	Parallelism  int                  // 同时扫描的最大 url 个数
	Pool         *ants.Pool           // 协程池，目前来看只是用来优化被动扫描，减小被动扫描时的协程创建、销毁的开销
	WG           sync.WaitGroup       // 等待协程池所有任务结束
	ScanTask     map[string]*ScanTask // 存储对目标扫描时的一些状态
	// Lock         sync.Mutex           // 对 Distribution函数中的一些 map 并发操作进行保护
	WgLock    deadlock.Mutex // ScanTask 是一个 map，运行插件时会并发操作，加锁保护
	WgAddLock deadlock.Mutex // ScanTask 是一个 map，运行插件时会并发操作，加锁保护
}

type ScanTask struct {
	PerServer map[string]bool                // 判断当前目标的 web server 是否扫过  key 为插件名字
	PerFolder map[string]bool                // 判断当前目标的目录是否扫过    这里的 key 为插件名_目录名 比如 bbscan_/admin
	PocPlugin map[string]bool                // 用来 poc 漏洞模块对应的指纹扫描是否扫，poc 模块依托于指纹识别，只有识别到了才会扫描
	Client    *httpx.Client                  // 用来进行请求的 client
	Archive   bool                           // 用来判断是否扫描过
	Wg        *sizedwaitgroup.SizedWaitGroup // 限制对每个url扫描时同时运行的插件数
}

// lock 对一些并发操作加锁
var lock sync.Mutex
var rex = regexp.MustCompile(`//#\s+sourceMappingURL=(.*\.map)`)

var seenRequests sync.Map // 这里主要是为了一些返回包检测类的判断是否识别过，减小开销，扫描类内部会判断是否扫描过

// DistributionTaskFunc ants 提交任务需要一个无参数的函数
type DistributionTaskFunc func()

// Distribution 对爬虫结果或者被动发现结果进行任务分发
func (t *Task) Distribution(in *input.CrawlResult) DistributionTaskFunc {
	return func() {
		if !conf.NoProgressBar {
			atomic.AddInt64(&output.TaskCounter, 1)
		}

		defer func() {
			t.WG.Done()
			logging.Logger.Debugln("扫描任务结束:", in.Url)
			if !conf.NoProgressBar {
				atomic.AddInt64(&output.TaskCompletionCounter, 1)
			}
		}()

		logging.Logger.Debugln(fmt.Sprintf("[%s] [%s] %s 扫描任务开始", in.UniqueId, in.Method, in.Url))
		// 这些返回包内容检测、指纹识别等因为没有使用检测是否扫描的逻辑，所以会重复检测，造成一定程度的资源消耗，问题应该不大
		// TODO 还没有想好怎么写逻辑，因为一些扫描插件会用到这些结果，搞成插件化的话，就需要控制插件的执行顺序，后续看看吧，目前影响不大
		msg := output.SCopilotData{
			Target:  in.Host,
			SiteMap: []string{in.Url},
		}

		// 过滤一些后缀, 比如 mp4 等，但 .css .js 还是要放过的，要进行敏感信息检测
		if strings.Contains(in.ContentType, "application/octet-stream") || strings.Contains(in.ContentType, "image/") || strings.Contains(in.ContentType, "video/") || strings.Contains(in.ContentType, "audio/") || funk.Contains(conf.GlobalConfig.Mitmproxy.FilterSuffix, filepath.Ext(in.ParseUrl.Path)) {
			// 更新数据
			output.SCopilot(in.Host, msg)
			return
		}

		var hostNoPort string
		_host := strings.Split(in.Host, ":")
		if len(_host) > 1 {
			hostNoPort = _host[0]
		} else {
			hostNoPort = in.Host
		}

		// 并发修改 in t.ScanTask，加锁保证安全, 使用全局唯一锁
		lock.Lock()
		if t.ScanTask[in.Host] == nil {
			t.ScanTask[in.Host] = &ScanTask{
				PerServer: make(map[string]bool),
				PerFolder: make(map[string]bool),
				PocPlugin: make(map[string]bool),
				Client:    httpx.NewClient(nil),
				// 3: 同时运行 3 个插件，3 供 PreServer、 PreFolder、PerFile这三个函数使用，防止马上退出 所以这里同时运行的插件个数为3-5 个
				// TODO 更优雅的实现方式
				Wg: sizedwaitgroup.New(3 + 3),
			}
		}

		// cdn 只检测一次
		if !t.ScanTask[in.Host].PerServer["cdnCheck"] {
			t.ScanTask[in.Host].PerServer["cdnCheck"] = true
			output.Lock.Lock()
			if _, ok := output.IPInfoList[hostNoPort]; !ok {
				// cdn 检测
				logging.Logger.Debugln("CheckCdn", hostNoPort)
				matched, value, itemType, dnsData := util.CheckCdn(hostNoPort)
				var ip string
				var allRecords []string
				var cdn bool
				if dnsData != nil {
					ip = strings.Join(dnsData.A, " ")
					if len(dnsData.A) > 1 { // 解析出多个 ip ，认为是 cdn
						cdn = true
					}
					allRecords = dnsData.AllRecords
					in.Ip = ip
				} else { // 这里说明传入的就是 ip
					in.Ip = hostNoPort
				}
				output.IPInfoList[hostNoPort] = &output.IPInfo{
					Ip:         ip,
					AllRecords: allRecords,
					Type:       itemType,
					Value:      value,
					Cdn:        matched,
				}
				if itemType == "cdn" || cdn {
					in.Cdn = true
				}
			}
			output.Lock.Unlock()
		}
		lock.Unlock()

		msg.HostNoPort = hostNoPort

		// 界面颜色显示 匹配
		go Marker(in.Url, in.RequestBody, in.Resp.Body, in.RawRequest, in.RawResponse, in.Id)

		// 指纹识别, 插件扫描会用到, 还是放到这里吧，多次识别也不影响
		fingerprints := fingprints.Identify([]byte(in.Resp.Body), in.Resp.RespHeader)
		if len(fingerprints) > 0 {
			msg.Fingerprints = append(msg.Fingerprints, fingerprints...)
			lock.Lock()
			in.Fingerprints = util.RemoveDuplicateElement(append(in.Fingerprints, msg.Fingerprints...))
			lock.Unlock()
		}

		msg.CollectionMsg = collection.Info(in.Url, hostNoPort, in.Resp.Body, in.ContentType)

		// 请求头中存在 Authorization 时，看看是不是 JWT ，如果是，自动对 JWT 进行爆破
		if v, ok := in.Headers["Authorization"]; ok {
			value := strings.Split(v, " ") // 有的 JWT ，是 Xxx Jwt 这种格式
			var jwtString string
			if len(value) > 1 {
				jwtString = value[1]
			} else {
				jwtString = value[0]
			}
			// 一个 twj 解析爆破一次就好了
			if !jwt.Jwts[jwtString] {
				// 首先解析一下，看看是不是 Jwt
				_, err := jwt.ParseJWT(jwtString)
				if err == nil { // 没有报错，说明解析 Jwt 成功
					msg.Fingerprints = util.RemoveDuplicateElement(append(msg.Fingerprints, "twj"))
					lock.Lock()
					in.Fingerprints = util.RemoveDuplicateElement(append(in.Fingerprints, msg.Fingerprints...))
					lock.Unlock()
					jwt.Jwts[jwtString] = true
					secret := jwt.GenerateSignature()
					if secret != "" {
						output.OutChannel <- output.VulMessage{
							DataType: "web_vul",
							Plugin:   "JWT",
							VulnData: output.VulnData{
								CreateTime: time.Now().Format("2006-01-02 15:04:05"),
								Target:     in.Url,
								Method:     in.Method,
								Ip:         in.Ip,
								Payload:    secret,
							},
							Level: output.Critical,
						}
					}
				}
			}
		}

		// 没有检测到 shiro 时，扫描的请求中添加一个 请求头检测一下 Cookie: rememberMe=3
		// TODO 这里不对，添加到这里，返回的数据包并没有经过这里，放到 httpx.request(...) 中 进行指纹检测的话，怎么返回获取指纹？还是说对于这种指纹检测，主动进行一次发包 (不太想使用这种方式)
		if !util.InSliceCaseFold("shiro", in.Fingerprints) {
			lock.Lock()
			if in.Headers["Cookie"] != "" {
				in.Headers["Cookie"] = in.Headers["Cookie"] + ";rememberMe=3"
			} else {
				in.Headers["Cookie"] = "rememberMe=3"
			}
			lock.Unlock()
		}

		// 更新数据
		output.SCopilot(in.Host, msg)

		errorMsg := sensitive.PageErrorMessageCheck(in.Url, in.RawRequest, in.Resp.Body)
		if len(errorMsg) > 0 {
			var res []string
			for _, v := range errorMsg {
				msg.Fingerprints = append(msg.Fingerprints, v.Type)
				res = append(res, v.Text)
			}
			msg.PluginMsg = []output.PluginMsg{
				{
					Url:      in.Url,
					Result:   res,
					Request:  in.RawRequest,
					Response: in.RawResponse,
				},
			}
		}

		// Reference 绕过测试，因为流量发的少，这个就不用插件控制了
		if (in.Resp.StatusCode == 403 || in.Resp.StatusCode == 401) && !isScanned(in.UniqueId) {
			bypass403.Reference(in.Target, in.Method, in.Headers, in.RequestBody, t.ScanTask[in.Host].Client)
		}

		// 非 css、js 类进行扫描, 单独进行 判断是否识别过
		if !strings.HasSuffix(in.ParseUrl.Path, ".css") && !strings.HasSuffix(in.ParseUrl.Path, ".js") {
			// 插件扫描
			t.Run(in)
			lock.Lock()
			// 收集参数
			paramNames, err := util.GetReqParameters(in.Method, in.ContentType, in.ParseUrl, []byte(in.RequestBody), in.Headers)

			if err != nil && !strings.Contains(err.Error(), "invalid semicolon separator in query") {
				logging.Logger.Errorln("GetReqParameters err:", err, in.Url, in.Method, in.ContentType, in.ParseUrl.Query(), in.RequestBody)
			} else {
				in.ParamNames = paramNames
				// 看请求、返回包中的参数是否包含敏感参数
				scan.PerFilePlugins["SensitiveParameters"].Scan("", "", in, nil)
				output.Lock.Lock()
				if output.SCopilotMessage[in.Host].CollectionMsg.Parameters == nil {
					output.SCopilotMessage[in.Host].CollectionMsg.Parameters = orderedmap.New()
				}
				resParamNames, _ := util.GetResParameters(strings.ToLower(in.Resp.RespHeader.Get("Content-Type")), []byte(in.Resp.Body))
				paramNames = append(paramNames, resParamNames...)
				for _, _para := range paramNames {
					v, ok := output.SCopilotMessage[in.Host].CollectionMsg.Parameters.Get(_para)
					if ok {
						output.SCopilotMessage[in.Host].CollectionMsg.Parameters.Set(_para, v.(int)+1)
					} else {
						output.SCopilotMessage[in.Host].CollectionMsg.Parameters.Set(_para, 1)
					}
				}
				// 按照value的字典序升序排序
				output.SCopilotMessage[in.Host].CollectionMsg.Parameters.Sort(func(a *orderedmap.Pair, b *orderedmap.Pair) bool {
					return a.Value().(int) > b.Value().(int)
				})
				output.Lock.Unlock()
			}

			// poc 模块依托于指纹识别，只有识别到对应的指纹才会扫描，所以这里就不插件化了
			if conf.Plugin["poc"] {
				t.ScanTask[in.Host].PocPlugin = pocs_go.PocCheck(in.Fingerprints, in.Target, in.Url, in.Ip, t.ScanTask[in.Host].PocPlugin, t.ScanTask[in.Host].Client)
			}
			lock.Unlock()
		} else {
			// 下面这些使用去重逻辑，因为扫描结果不会被别的插件用到
			if isScanned(in.UniqueId) {
				return
			}

			if strings.HasSuffix(in.ParseUrl.Path, ".js") {
				// 对于 js 这种单独判断是否扫描过，减少消耗
				if strings.HasPrefix(in.Resp.Body, "webpackJsonp(") || strings.Contains(in.Resp.Body, "window[\"webpackJsonp\"]") {
					msg.Fingerprints = util.RemoveDuplicateElement(append(msg.Fingerprints, "Webpack"))
					lock.Lock()
					in.Fingerprints = util.RemoveDuplicateElement(append(in.Fingerprints, msg.Fingerprints...))
					lock.Unlock()
				}

				// 前端 js 中存在 sourcemap 文件，即 xxx.js.map 这种可以使用 sourcemap 等工具还原前端代码
				match := rex.FindStringSubmatch(in.Resp.Body)
				if match != nil {
					msg.Fingerprints = util.RemoveDuplicateElement(append(msg.Fingerprints, "SourceMap"))
					lock.Lock()
					in.Fingerprints = util.RemoveDuplicateElement(append(in.Fingerprints, msg.Fingerprints...))
					lock.Unlock()
					output.OutChannel <- output.VulMessage{
						DataType: "web_vul",
						Plugin:   "SourceMap",
						VulnData: output.VulnData{
							CreateTime: time.Now().Format("2006-01-02 15:04:05"),
							Target:     in.Url,
						},
						Level: output.Low,
					}
				}
			}
		}

		// 更新数据
		output.SCopilot(in.Host, msg)

		go func() {
			// 判断 Archive 是否有数据，如果有的话分发至扫描
			if !t.ScanTask[in.Host].Archive && in.Archive != nil {
				for k, v := range in.Archive {
					response, err := t.ScanTask[in.Host].Client.Request(v, "GET", "", nil, "Archive")
					if err != nil {
						continue
					}
					parseUrl, err := url.Parse(v)
					if err != nil {
						continue
					}

					if response.StatusCode == 200 && !util.IsBlackHtml(response.Body, response.RespHeader["Content-Type"], parseUrl.Path) {
						i := &input.CrawlResult{
							Target:   in.Target,
							Url:      v,
							Host:     in.Host,
							ParseUrl: parseUrl,
							UniqueId: k,
							Method:   "GET",
							Resp: &httpx.Response{
								Status:     strconv.Itoa(response.StatusCode),
								StatusCode: response.StatusCode,
								Body:       response.Body,
								RespHeader: response.RespHeader,
							},
							RawRequest:  response.RequestDump,
							RawResponse: response.ResponseDump,
						}
						t.Run(i)
					}
				}
			}
			lock.Lock()
			t.ScanTask[in.Host].Archive = true
			lock.Unlock()
		}()
	}
}

func isScanned(key string) bool {
	if key == "" {
		return false
	}
	if _, ok := seenRequests.Load(key); ok {
		return true
	}
	seenRequests.Store(key, true)
	return false
}
