package pkg

import (
    "fmt"
    "github.com/yhy0/ChYing/lib/webUnPack/pkg/api"
    "github.com/yhy0/ChYing/lib/webUnPack/pkg/output"
    "github.com/yhy0/ChYing/lib/webUnPack/pkg/sourcemap"
    "github.com/yhy0/logging"
    "sync"
)

/**
  @author: yhy
  @since: 2023/7/26
  @desc: //TODO
**/

type Options struct {
    Target string
    Output string
}

func Run(options *Options) {
    c := collector.Collector{}
    c.Workers = 20
    
    if options.Output != "" {
        c.Output = options.Output
    }
    
    logging.Logger.Infoln("running", options.Target)
    
    wg := sync.WaitGroup{}
    c.Init()
    wg.Add(1)
    go func() {
        defer wg.Done()
        c.Run()
    }()
    
    go func() {
        for result := range output.ResultChan {
            logging.Logger.Debugf("[OutPut] [Type]: %s, [Value]: %s, [Source]: %s", result.Type, result.Value, result.Source)
            switch result.Type {
            case "api":
                api.Enumerate(c.Target, result)
            case "domain":
                fmt.Println(result.Value)
            }
        }
    }()
    
    err := c.Add(options.Target)
    
    c.Close()
    
    if err != nil {
        logging.Logger.Errorln("cannot emit urls", err)
    }
    wg.Wait()
    
    logging.Logger.Infoln("finished")
}
