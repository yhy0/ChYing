package nucleiY

import (
	"bufio"
	"context"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/nuclei/v2/pkg/catalog/disk"
	"github.com/projectdiscovery/nuclei/v2/pkg/catalog/loader"
	"github.com/projectdiscovery/nuclei/v2/pkg/core"
	"github.com/projectdiscovery/nuclei/v2/pkg/output"
	"github.com/projectdiscovery/nuclei/v2/pkg/parsers"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/hosterrorscache"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/interactsh"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/protocolinit"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols/common/protocolstate"
	"github.com/projectdiscovery/nuclei/v2/pkg/reporting"
	"github.com/projectdiscovery/nuclei/v2/pkg/templates"
	"github.com/projectdiscovery/nuclei/v2/pkg/testutils"
	"github.com/projectdiscovery/nuclei/v2/pkg/types"
	"github.com/projectdiscovery/ratelimit"
	errorutil "github.com/projectdiscovery/utils/errors"
	fileutil "github.com/projectdiscovery/utils/file"
	proxyutils "github.com/projectdiscovery/utils/proxy"
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/logging"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/**
   @author yhy
   @since 2023/6/8
   @desc //TODO
**/

type Nuclei struct {
	Engine *core.Engine
	Store  *loader.Store
}

var ResultEvent = make(chan *output.ResultEvent)

var nuclei = &Nuclei{}

func New(proxy string) {
	templatesTempDir := filepath.Join(file.ChyingDir, "nucleiY")

	if _, err := os.Stat(templatesTempDir); err != nil {
		// 不存在，创建
		logging.Logger.Errorln("nucleiY not find")
		return
	}

	cache := hosterrorscache.New(30, hosterrorscache.DefaultMaxHostsCount, nil)
	defer cache.Close()

	mockProgress := &testutils.MockProgressClient{}
	reportingClient, _ := reporting.New(&reporting.Options{}, "")
	defer reportingClient.Close()

	defaultOpts := types.DefaultOptions()
	protocolstate.Init(defaultOpts)
	protocolinit.Init(defaultOpts)

	// 模板路径是通过这个指定的
	defaultOpts.Templates = goflags.StringSlice{templatesTempDir}
	if proxy != "" {
		defaultOpts.Proxy = goflags.StringSlice{proxy}
		if err := loadProxyServers(defaultOpts); err != nil {
			logging.Logger.Errorln(err)
		}
	}
	outputWriter := testutils.NewMockOutputWriter()
	outputWriter.WriteCallback = func(event *output.ResultEvent) {
		// 这样写，不能 单独 这样 ResultEvent <- event ，不然会阻塞
		select {
		case ResultEvent <- event:
		default:
		}
	}

	interactOpts := interactsh.DefaultOptions(outputWriter, reportingClient, mockProgress)
	interactClient, err := interactsh.New(interactOpts)
	if err != nil {
		logging.Logger.Fatalf("Could not create interact client: %s\n", err)
	}
	defer interactClient.Close()

	catalog := disk.NewCatalog("")
	executerOpts := protocols.ExecutorOptions{
		Output:          outputWriter,
		Options:         defaultOpts,
		Progress:        mockProgress,
		Catalog:         catalog,
		IssuesClient:    reportingClient,
		RateLimiter:     ratelimit.New(context.Background(), 150, time.Second),
		Interactsh:      interactClient,
		HostErrorsCache: cache,
		Colorizer:       aurora.NewAurora(true),
		ResumeCfg:       types.NewResumeCfg(),
	}
	engine := core.New(defaultOpts)
	engine.SetExecuterOptions(executerOpts)

	workflowLoader, err := parsers.NewLoader(&executerOpts)
	if err != nil {
		logging.Logger.Fatalf("Could not create workflow loader: %s\n", err)
	}
	executerOpts.WorkflowLoader = workflowLoader

	store, err := loader.New(loader.NewConfig(defaultOpts, catalog, executerOpts))
	if err != nil {
		logging.Logger.Fatalf("Could not create loader client: %s\n", err)
	}
	store.Load()

	Pocs = make(map[string][]*templates.Template)

	for _, t := range store.Templates() {
		for _, tag := range t.Info.Tags.ToSlice() {
			if tag != "" {
				value, ok := Pocs[tag]
				if ok {
					Pocs[tag] = append(value, t)
				} else {
					Pocs[tag] = []*templates.Template{t}
				}
			}
		}

	}

	nuclei.Engine = engine
	nuclei.Store = store
}

// loadProxyServers load list of proxy servers from file or comma seperated
func loadProxyServers(options *types.Options) error {
	if len(options.Proxy) == 0 {
		return nil
	}
	proxyList := []string{}
	for _, p := range options.Proxy {
		if fileutil.FileExists(p) {
			file, err := os.Open(p)
			if err != nil {
				return fmt.Errorf("could not open proxy file: %w", err)
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				proxy := scanner.Text()
				if strings.TrimSpace(proxy) == "" {
					continue
				}
				proxyList = append(proxyList, proxy)
			}
		} else {
			proxyList = append(proxyList, p)
		}
	}
	aliveProxy, err := proxyutils.GetAnyAliveProxy(options.Timeout, proxyList...)
	if err != nil {
		return err
	}
	proxyURL, err := url.Parse(aliveProxy)
	if err != nil {
		return errorutil.WrapfWithNil(err, "failed to parse proxy got %v", err)
	}
	if options.ProxyInternal {
		os.Setenv(types.HTTP_PROXY_ENV, proxyURL.String())
	}
	if proxyURL.Scheme == proxyutils.HTTP || proxyURL.Scheme == proxyutils.HTTPS {
		types.ProxyURL = proxyURL.String()
		types.ProxySocksURL = ""
		logging.Logger.Infof("Using %s as proxy server", proxyURL.String())
	} else if proxyURL.Scheme == proxyutils.SOCKS5 {
		types.ProxyURL = ""
		types.ProxySocksURL = proxyURL.String()
		logging.Logger.Infof("Using %s as socket proxy server", proxyURL.String())
	}
	return nil
}
