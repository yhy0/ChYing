package sqlmap

/**
   @author yhy
   @since 2023/10/27
   @desc //TODO
**/

// http://127.0.0.1:8775/option/taskid/list
type option struct {
    Url         string `json:"url"`
    Method      string `json:"method"`
    Headers     string `json:"headers"`
    Data        string `json:"data"`
    RandomAgent bool   `json:"randomAgent"`
    Level       int    `json:"level"`
    Risk        int    `json:"risk"`
    Proxy       string `json:"proxy"`
    Verbose     int    `json:"verbose"`
}
