package mode

import (
    "github.com/panjf2000/ants/v2"
    "github.com/yhy0/ChYing/mitmproxy"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/task"
    "github.com/yhy0/logging"
)

/**
  @author: yhy
  @since: 2023/1/11
  @desc: 被动代理数据处理
**/

func Passive() {
    logging.Logger.Debugln("Start passive traffic monitoring scan")
    pool, _ := ants.NewPool(conf.Parallelism + 1)
    t := &task.Task{
        Parallelism: conf.Parallelism + 1,
        ScanTask:    make(map[string]*task.ScanTask),
        Pool:        pool,
    }
    
    go mitmproxy.NewPassiveTask(t)
}
