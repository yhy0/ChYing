package output

import (
    "fmt"
    "github.com/logrusorgru/aurora"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "time"
)

/**
   @author yhy
   @since 2023/11/8
   @desc 处理进度
**/

var TaskCounter int64
var TaskCompletionCounter int64

func Progress() {
    if conf.NoProgressBar {
        return
    }
    i := 0
    for {
        if TaskCounter == TaskCompletionCounter {
            i++
            if i == 10 {
                i = 0
                fmt.Println(aurora.Yellow(fmt.Sprintf("A total of %v tasks have been received, %v have been processed, and the processing rate is %.2f", TaskCounter, TaskCompletionCounter, float64(TaskCompletionCounter)/float64(TaskCounter)*100) + "%").String())
            }
            time.Sleep(10 * time.Second)
            continue
        } else {
            fmt.Println(aurora.Yellow(fmt.Sprintf("A total of %v tasks have been received, %v have been processed, and the processing rate is %.2f", TaskCounter, TaskCompletionCounter, float64(TaskCompletionCounter)/float64(TaskCounter)*100) + "%").String())
            time.Sleep(5 * time.Second)
        }
    }
}
