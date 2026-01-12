package collector

import (
    "github.com/yhy0/logging"
    "net/url"
    "sync"
)

type Task interface {
    Name() string
    URLs() <-chan *url.URL
    Run(*url.URL) error
    Finish()
}

type SourceMap struct {
    FileNames []string `json:"sources"`
    Contents  []string `json:"sourcesContent"`
}

type Collector struct {
    Output  string
    Workers int
    Target  string
    pages   chan *url.URL
}

func (c *Collector) Init() {
    c.pages = make(chan *url.URL)
}

func (c *Collector) Add(purl string) error {
    parsed, err := url.Parse(purl)
    if err != nil {
        return err
    }
    c.Target = parsed.String()
    c.pages <- parsed
    return nil
}

func (c *Collector) Close() {
    close(c.pages)
}

func (c *Collector) Run() {
    wg := sync.WaitGroup{}
    scripts := make(chan *url.URL)
    maps := make(chan *url.URL)
    infos := make(chan *url.URL)
    
    c.spawn(&wg, c.runWorkers, &TaskPages{
        In:  c.pages,
        Out: scripts,
    })
    c.spawn(&wg, c.runWorkers, &TaskScripts{
        Output:  c.Output,
        In:      scripts,
        Out:     maps,
        Visited: make(map[string]struct{}),
        Mutex:   &sync.Mutex{},
        Target:  c.Target,
    })
    c.spawn(&wg, c.runWorkers, &TaskMaps{
        Output: c.Output,
        In:     maps,
        Out:    infos,
        Target: c.Target,
    })
    c.worker(&TaskInfos{
        Output: c.Output,
        In:     infos,
    })
    
    wg.Wait()
}

func (Collector) spawn(wg *sync.WaitGroup, fn func(Task), task Task) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        fn(task)
    }()
}

func (c *Collector) runWorkers(task Task) {
    wg := sync.WaitGroup{}
    for i := 0; i < c.Workers; i++ {
        c.spawn(&wg, c.worker, task)
    }
    wg.Wait()
    task.Finish()
}

func (c *Collector) worker(task Task) {
    for url := range task.URLs() {
        logging.Logger.Debugln("task running", task.Name(), url.String())
        err := task.Run(url)
        if err != nil {
            logging.Logger.Errorln("task error", err, task.Name(), url.String())
        }
    }
}
