package httpx

import "regexp"

/**
  @author: yhy
  @since: 2023/5/24
  @desc: //TODO
**/

func GetTitle(body string) string {
	titleReg := regexp.MustCompile(`<title>([\s\S]{1,200})</title>`)
	title := titleReg.FindStringSubmatch(body)
	if len(title) > 1 {
		return title[1]
	}
	return ""
}
