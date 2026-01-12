package collector

import (
    "fmt"
    "net/url"
    "os"
    "path"
)

// TaskInfos consumes source maps URLs and saves them into _sources.txt file.
type TaskInfos struct {
    Output string
    In     chan *url.URL
}

func (TaskInfos) Name() string {
    return "infos"
}

func (task *TaskInfos) Finish() {}

func (task *TaskInfos) URLs() <-chan *url.URL {
    return task.In
}

func (task *TaskInfos) Run(surl *url.URL) error {
    if task.Output != "" {
        fpath := path.Join(task.Output, surl.Hostname(), "_sources.txt")
        f, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
        if err != nil {
            return fmt.Errorf("open _sources.txt: %v", err)
        }
        defer f.Close()
        _, err = f.WriteString(surl.String() + "\n")
        if err != nil {
            return fmt.Errorf("write _sources.txt: %v", err)
        }
    }
    
    return nil
}
