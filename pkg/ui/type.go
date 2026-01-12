package ui

/**
   @author yhy
   @since 2024/12/16
   @desc //TODO
**/

type Result struct {
    Url           string `json:"url"`
    Method        string `json:"method"`
    StatusCode    int    `json:"status"`
    ContentLength int    `json:"length"`
    Request       string `json:"request"`
    Response      string `json:"response"`
}
