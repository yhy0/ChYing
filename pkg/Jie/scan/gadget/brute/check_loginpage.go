package brute

import (
	"github.com/thoas/go-funk"
	regexp "github.com/wasilibs/go-re2"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"net/url"
	"strings"
)

func CheckLoginPage(inputurl string, body string, client *httpx.Client) bool {
	client.Options.AllowRedirect = 5
	if body == "" {
		req, err := client.Request(inputurl, "GET", "", nil, "Brute")
		if err != nil {
			return false
		}
		body = req.Body
	}
	if funk.Contains(body, "<title> Login </title>") {
		return true
	}
	cssurl := regexp.MustCompile(`<link[^>]*href=['"](.*?)['"]`).FindAllStringSubmatch(body, -1)
	for _, v := range cssurl {
		if strings.Contains(v[1], ".css") {
			u, err := url.Parse(inputurl)
			if err != nil {
				return false
			}
			href, err := url.Parse(v[1])
			if err != nil {
				return false
			}
			if err != nil {
				return false
			}
			hrefurl := u.ResolveReference(href)
			if reqcss, err := client.Request(hrefurl.String(), "GET", "", nil, "Brute"); err == nil {
				if strings.Contains(reqcss.Body, "login") || strings.Contains(reqcss.Body, "Login") {
					return true
				}
			}
		}

		return false
	}
	return false
}
