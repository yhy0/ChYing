package sql

import (
    "fmt"
    "github.com/sergi/go-diff/diffmatchpatch"
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/input"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
    "github.com/yhy0/ChYing/pkg/Jie/pkg/util"
    "github.com/yhy0/logging"
    "strconv"
    "sync"
    "testing"
)

/**
  @author: yhy
  @since: 2023/2/11
  @desc: //TODO
**/

func TestSqlmap(t *testing.T) {
    logging.Logger = logging.New(true, "", "", true)
    conf.GlobalConfig = &conf.Config{}
    conf.GlobalConfig.Http.Proxy = "http://127.0.0.1:8080"
    
    // 使用 sync.WaitGroup 防止 OutChannel 中的数据没有完全被消费，导致的数据漏掉问题
    var wg sync.WaitGroup
    wg.Add(1)
    
    go func() {
        defer wg.Done()
        for v := range output.OutChannel {
            fmt.Println(v.PrintScreen())
        }
    }()
    
    in := &input.CrawlResult{
        Url:         "http://127.0.0.1/Less-1/?id=2",
        Method:      "GET",
        RequestBody: "",
        Param:       []string{"id"},
        Headers:     map[string]string{},
    }
    
    response, err := httpx.Request(in.Url, in.Method, in.RequestBody, nil)
    if err != nil {
        return
    }
    
    in.Resp = response
    sqlPlugin := &Plugin{}
    client := httpx.NewClient(nil)
    sqlPlugin.Scan(in.Url, "", in, client)
    close(output.OutChannel)
    wg.Wait()
}

func TestDiff(t *testing.T) {
    text1 := "Lorem ipsum dolor sit amet, congue tation referrentur ei sed. Ne nec legimus habemus recusabo, natum reque et per. <script src='ads.js'></script>Facer tritani reprehendunt eos id, modus constituam est te. Usu sumo indoctum ad, pri paulo molestiae complectitur no."
    text2 := "Lorem ipsum dolor sit amet, congue tation referrentur ei sed. Ne nec legimus habemus recusabo, natum reque et per. Facer tritani reprehendunt eos id, modus constituam est te. Usu sumo indoctum ad, pri paulo molestiae complectitur no."
    
    text1 = "Lorem dolor ipsum asds."
    text2 = "Lorem dolor sit amet."
    
    dmp := diffmatchpatch.New()
    
    diffs := dmp.DiffMain(text1, text2, true)
    
    // 这里简单粗暴
    for _, diff := range diffs {
        
        switch diff.Type {
        case diffmatchpatch.DiffInsert:
            fmt.Println(" [DiffInsert] ", diff.Text)
        
        case diffmatchpatch.DiffDelete:
            fmt.Println(" [DiffDelete] ", diff.Text)
            // case diffmatchpatch.DiffEqual:
            //    _, _ = buff.WriteString(diff.Text)
            // }
        }
        
    }
}

func TestDiff1(t *testing.T) {
    text1 := "Lorem ipsum dolor sit amet, congue tation referrentur ei sed. Ne nec legimus habemus recusabo, natum reque et per. Facer tritani reprehendunt eos id, modus constituam est te. Usu sumo indoctum ad, pri paulo molestiae complectitur no."
    text2 := "Lorem ipsum dolor sit amet, congue tation referrentur ei sed. Ne nec legimus habemus recusabo, natum reque et per. <script src='ads.js'></script>Facer tritani reprehendunt eos id, modus constituam est te. Usu sumo indoctum ad, pri paulo molestiae complectitur no."
    
    prefix, suffix := findDynamicContent(text1, text2)
    
    /*
       prefix: natum reque et per.
       suffix: Facer tritani repreh
    */
    fmt.Println(prefix)
    fmt.Println(suffix)
    
}

func TestRemove(t *testing.T) {
    text1 := "Lorem ipsum dolor sit amet, congue tation referrentur ei sed. Ne nec legimus habemus recusabo, natum reque et per. Facer tritani reprehendunt eos id, modus constituam est te. Usu sumo indoctum ad, pri paulo molestiae complectitur no."
    text2 := "Lorem ipsum dolor sit amet, congue tation referrentur ei sed. Ne nec legimus habemus recusabo, natum reque et per. <script src='ads.js'></script>Facer tritani reprehendunt eos id, modus constituam est te. Usu sumo indoctum ad, pri paulo molestiae complectitur no."
    
    var sql = Sqlmap{
        DynamicMarkings: map[string]string{
            "prefix": "",
            "suffix": "",
        },
    }
    
    prefix, suffix := findDynamicContent(text1, text2)
    sql.DynamicMarkings["prefix"] = prefix
    sql.DynamicMarkings["suffix"] = suffix
    
    /*
       输出内容，现在看是和 sql 一致，但没有经过大量测试，TODO 可能有待优化
       Lorem ipsum dolor sit amet, congue tation referrentur ei sed. Ne nec legimus habemus recusabo, natum reque et per. Facer tritani reprehendunt eos id, modus constituam est te. Usu sumo indoctum ad, pri paulo molestiae complectitur no.
    
    */
    fmt.Println(sql.removeDynamicContent(text2))
}

func TestPayload(t *testing.T) {
    randomTestString := getErrorBasedPreCheckPayload()
    fmt.Println(randomTestString)
    
    fmt.Println(randomTestString + strconv.Itoa(util.RandomNumber(3, 5)))
    
}
