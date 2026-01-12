package utils

import (
    "fmt"
    "github.com/yhy0/logging"
    "testing"
)

/**
   @author yhy
   @since 2024/9/2
   @desc //TODO
**/

func TestFuzzApi(t *testing.T) {
    logging.Logger = logging.New(true, "", "ChYing", true)
    apis := []string{
        "/v1/api/user",
        "/v1/api/name",
        "/v2/data",
    }
    fmt.Println(PredictionApi(apis, 1))
}
