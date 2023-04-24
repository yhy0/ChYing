package tools

/**
  @author: yhy
  @since: 2023/4/23
  @desc: //TODO
**/

type Result struct {
	Url           string `json:"url"`
	StatusCode    int    `json:"status"`
	ContentLength int    `json:"length"`
	Request       string `json:"request"`
	Response      string `json:"response"`
}
