package fuzz

import (
	"github.com/yhy0/ChYing/pkg/httpx"
	"github.com/yhy0/Jie/conf"
	"github.com/yhy0/Jie/pkg/util"
	"strings"
)

/**
  @author: yhy
  @since: 2023/4/23
  @desc: //TODO
**/

func filter(path string, resp *httpx.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	// 返回是个图片
	if util.Contains(contentType, "image/") {
		return true
	}

	if strings.HasSuffix(path, ".xml") {
		if !util.Contains(contentType, "xml") {

			return true
		}
	} else if strings.HasSuffix(path, ".json") {
		if !util.Contains(contentType, "json") {
			return true
		}
	}

	// 文件内容为空丢弃
	if resp.ContentLength == 0 {
		return true
	}

	title := getTitle(resp.Body)

	if util.In(title, conf.Page404Title) {
		return true
	}

	if util.In(resp.Body, conf.Page404Content) {
		return true
	}

	if util.In(resp.Body, conf.WafContent) {
		return true
	}

	return false
}
