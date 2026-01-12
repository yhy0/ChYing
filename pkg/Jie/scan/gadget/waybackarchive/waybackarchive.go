package waybackarchive

import (
	"fmt"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
	"strings"
)

/**
  @author: yhy
  @since: 2023/11/17
  @desc: 利用 https://web.archive.org/ 进行获取历史 url 链接(参数)，然后进行扫描
**/

func Run(u string, client *httpx.Client) map[string]string {
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}

	resp, err := client.Request(fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s*&output=txt&fl=original&collapse=urlkey&fastLatest=true", u), "GET", "", nil, "WayBackArchive")
	if err != nil {
		logging.Logger.Debugln("WayBackArchive err:", err)
		return nil
	}

	lines := strings.Split(resp.Body, "\n")

	// 收集到的 url 有很多只是参数不一样，所以这里进行判断，只要唯一的 url
	uniqueUrl := make(map[string]string)

	for _, line := range lines {
		if line == "" {
			continue
		}
		id := util.SimpleUniqueId(line)
		if id == "" {
			continue
		}

		if _, ok := uniqueUrl[id]; ok {
			continue
		}
		uniqueUrl[id] = line
	}

	return uniqueUrl
}
