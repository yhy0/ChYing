package nucleiY

import (
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
	"github.com/yhy0/ChYing/pkg/file"
	"github.com/yhy0/logging"
	"path/filepath"
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

var nuclei Nuclei

func New() {
	templatesTempDir := filepath.Join(file.ChyingDir, "nucleiY")

	cache := hosterrorscache.New(30, hosterrorscache.DefaultMaxHostsCount, nil)
	defer cache.Close()

	mockProgress := &testutils.MockProgressClient{}
	reportingClient, _ := reporting.New(&reporting.Options{}, "")
	defer reportingClient.Close()

	outputWriter := testutils.NewMockOutputWriter()
	outputWriter.WriteCallback = func(event *output.ResultEvent) {
		fmt.Printf("Got Result: %v\n", event)
	}

	defaultOpts := types.DefaultOptions()
	protocolstate.Init(defaultOpts)
	protocolinit.Init(defaultOpts)

	// 模板路径是通过这个指定的
	defaultOpts.Templates = goflags.StringSlice{templatesTempDir}

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

	for _, t := range store.Templates() {
		tag := t.Info.Tags.ToSlice()[0]
		value, ok := Pocs[tag]
		if ok {
			Pocs[tag] = append(value, t)
		} else {
			Pocs[tag] = []*templates.Template{t}
		}
	}

	nuclei.Engine = engine
	nuclei.Store = store
}
