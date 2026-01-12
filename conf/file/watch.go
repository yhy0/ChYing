package file

import (
    "github.com/fsnotify/fsnotify"
    "github.com/yhy0/logging"
    "path/filepath"
)

/**
   @author yhy
   @since 2024/12/30
   @desc 监控文件变动
**/

func watch() {
    // 创建一个新的文件系统监视器
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        logging.Logger.Fatal(err)
    }
    defer watcher.Close()
    
    // 监视配置文件
    mitmPath := filepath.Join(ChyingDir, "default_mitm_rule.json")
    err = watcher.Add(mitmPath)
    if err != nil {
        logging.Logger.Fatal(err)
    }
    
    // 保持程序运行
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            if event.Op&fsnotify.Write == fsnotify.Write {
                logging.Logger.Println("配置文件已修改:", event.Name)
                readMitmRule()
            }
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            logging.Logger.Println("错误:", err)
        }
    }
}
