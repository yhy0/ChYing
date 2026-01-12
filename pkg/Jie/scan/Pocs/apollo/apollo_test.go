package apollo

import (
    "github.com/yhy0/ChYing/pkg/Jie/conf"
    "github.com/yhy0/logging"
    "testing"
)

func TestApollo(t *testing.T) {
    logging.Logger = logging.New(true, "", "agent", true)
    conf.GlobalConfig = &conf.Config{}
    conf.GlobalConfig.Http.Proxy = "http://127.0.0.1:8080"
    
    Run("http://172.16.39.132:8091", "http://172.16.39.132:8089")
}
