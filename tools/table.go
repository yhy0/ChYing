package tools

/**
  @author: yhy
  @since: 2023/4/23
  @desc: //TODO
**/

type Result struct {
	Url           string `json:"url"`
	Method        string `json:"method"`
	StatusCode    int    `json:"status"`
	ContentLength int    `json:"length"`
	Request       string `json:"request"`
	Response      string `json:"response"`
}
