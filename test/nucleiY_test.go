package test

import (
	"fmt"
	"github.com/yhy0/ChYing/tools/nucleiY"
	"github.com/yhy0/logging"
	_ "net/http/pprof"
	"testing"
	"time"
)

/**
   @author yhy
   @since 2023/6/8
   @desc //TODO
**/

func TestNucleiY(t *testing.T) {

	logging.New(true, "", "nucleiY")

	go func() {
		for event := range nucleiY.ResultEvent {
			logging.Logger.Infof("%v", event)
		}
		//select {
		//case event := <-nucleiY.ResultEvent:
		//	logging.Logger.Infof("%v", event)
		//}
	}()

	nucleiY.Scan("http://127.0.0.1:18090/", "ecology:Ecology - Local File Inclusion", "")

	time.Sleep(1 * time.Second)

	nucleiY.Scan("http://127.0.0.1:18090/", "ecology-all", "http://127.0.0.1:8080")
	fmt.Println("--------------------")

}
