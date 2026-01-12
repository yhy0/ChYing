package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/thoas/go-funk"
	"github.com/yhy0/ChYing/pkg/Jie/pkg/util"
	"github.com/yhy0/logging"
	"io"

	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strings"
)

/*
*
    @author: yhy
    @since: 2023/1/30
    @desc: https://github.com/wrenchonline/glint/blob/master/util/util.go
*
*/

const (
	applicationJson       = "application/json"
	applicationUrlencoded = "application/x-www-form-urlencoded"
	multipartData         = "multipart/form-data"
	unknown               = "unknown"
)

// Param describes an individual posted parameter.
type Param struct {
	// Name of the posted parameter.
	Name string `json:"name"`
	// Value of the posted parameter.
	Value string `json:"value,omitempty"`
	// Filename of a posted file.
	Filename string `json:"fileName,omitempty"`
	// ContentType is the content type of posted file.
	ContentType string `json:"contentType,omitempty"`

	FileHeader   textproto.MIMEHeader
	FileSize     int64
	FileContent  []byte
	IsFile       bool
	Boundary     string
	FilenotFound bool
	IsBase64     bool
	Index        int //
}

type Variations struct {
	// MimeType is the MIME type of the posted data.
	MimeType string `json:"mimeType"`
	// Params is a list of posted parameters (in case of URL encoded parameters).
	Params []Param `json:"params"`
	// OriginalParams 存储原始的值
	OriginalParams []Param `json:"original_params"`
	// Text contains the posted data. Although its type is string, it may contain
	// binary data.
	Text string `json:"text"`
}

func (p Variations) Len() int {
	return len(p.Params)
}

// Less 顺序有低到高排序
func (p Variations) Less(i, j int) bool {
	return p.Params[i].Index < p.Params[j].Index
}

func (p Variations) Swap(i, j int) {
	p.Params[i], p.Params[j] = p.Params[j], p.Params[i]
}

// ParseUri 对请求进行格式化
func ParseUri(uri string, body []byte, method string, contentType string, headers map[string]string) (*Variations, error) {
	var (
		err        error
		index      int
		variations Variations
	)

	jsonMap := make(map[string]interface{})
	if strings.ToUpper(method) == "POST" {
		if len(body) > 0 {
			switch getContentType(strings.ToLower(contentType)) {
			case applicationJson:
				err := json.Unmarshal(body, &jsonMap)
				if err != nil {
					return nil, err
				}
				for k, v := range jsonMap {
					index++
					if v != nil {
						if value, ok := v.(string); ok {
							Post := Param{Name: k, Value: value, Index: index, ContentType: contentType}
							variations.Params = append(variations.Params, Post)
							variations.OriginalParams = append(variations.OriginalParams, Post)
						}
					}
				}
				variations.MimeType = applicationJson
			case multipartData:
				var iindex = 0
				var boundary string
				// base64.StdEncoding.DecodeString (
				// array := string()
				iobody := bytes.NewReader(body)
				req, err := http.NewRequest(method, uri, iobody)

				for k, v := range headers {
					req.Header[k] = []string{v}
				}

				if err != nil {
					logging.Logger.Error(err.Error())
					return nil, err
				}

				reader, err := req.MultipartReader()
				if err != nil {
					logging.Logger.Error(err.Error())
					return nil, err
				}

				_, params, err := mime.ParseMediaType(contentType)
				if err != nil {
					logging.Logger.Error("mime.ParseMediaType :", err)
				}
				if value, ok := params["boundary"]; ok {
					boundary = value
				}

				for {
					var isfile = false
					if reader == nil {
						break
					}
					p, err := reader.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						logging.Logger.Error("mime.ParseMediaType :", err)
						return nil, err
					}

					body, err := io.ReadAll(p)
					if err != nil {
						p.Close()
						return nil, err
					}
					iindex++
					if p.FileName() != "" {
						isfile = true
					}

					variations.MimeType = multipartData
					variations.Params = append(variations.Params, Param{
						Name:        p.FormName(),
						Boundary:    boundary,
						Filename:    p.FileName(),
						ContentType: p.Header.Get("Content-Type"),
						// FileContent: body,
						Value:  string(body),
						IsFile: isfile,
						Index:  iindex,
					})

					variations.OriginalParams = append(variations.OriginalParams, Param{
						Name:        p.FormName(),
						Boundary:    boundary,
						Filename:    p.FileName(),
						ContentType: p.Header.Get("Content-Type"),
						// FileContent: body,
						Value:  string(body),
						IsFile: isfile,
						Index:  iindex,
					})

					p.Close()
				}
			default:
				strs := strings.Split(string(body), "&")
				for i, kv := range strs {
					kvs := strings.Split(kv, "=")
					if len(kvs) == 2 {
						key := kvs[0]
						value := kvs[1]
						Post := Param{Name: key, Value: value, Index: i, ContentType: contentType}
						variations.Params = append(variations.Params, Post)
						variations.OriginalParams = append(variations.OriginalParams, Post)
					} else {
						return nil, fmt.Errorf("%s exec function strings.Split fail", uri)
					}
				}
				variations.MimeType = contentType
			}
			if &variations == nil {
				return nil, fmt.Errorf("%s variations is nil", method)
			}
			sort.Sort(variations)
			return &variations, nil
		} else {
			return nil, fmt.Errorf("%s POST data is empty", uri)
		}

	} else if strings.ToUpper(method) == "GET" {
		if !funk.Contains(uri, "?") {
			return nil, fmt.Errorf("%s GET data is empty", uri)
		}
		urlparams := strings.TrimRight(uri, "&")
		urlparams = strings.Split(uri, "?")[1]
		strs := strings.Split(urlparams, "&")
		for i, kv := range strs {
			kvs := strings.Split(kv, "=")
			if len(kvs) == 2 {
				key := kvs[0]
				value := kvs[1]
				variations.Params = append(variations.Params, Param{Name: key, Value: value, Index: i, ContentType: contentType})
				variations.OriginalParams = append(variations.OriginalParams, Param{Name: key, Value: value, Index: i, ContentType: contentType})
			} else {
				logging.Logger.Errorln("exec function strings.Split fail, ", uri)
				continue
			}
		}
		sort.Sort(variations)
		if &variations == nil {
			return nil, fmt.Errorf("%s variations is nil", method)
		}
		return &variations, nil

	} else {
		err = fmt.Errorf("%s method not supported", method)
	}
	return nil, err
}

func getContentType(data string) string {
	if funk.Contains(data, applicationJson) {
		return applicationJson
	}
	if funk.Contains(data, applicationUrlencoded) {
		return applicationUrlencoded
	}
	if funk.Contains(data, multipartData) {
		return multipartData
	}
	return unknown
}

func (p *Variations) Release() string {
	var buf bytes.Buffer
	mjson := make(map[string]interface{})
	if p.MimeType == applicationJson {
		for _, param := range p.Params {
			mjson[param.Name] = param.Value
		}
		jsonary, err := json.Marshal(mjson)
		if err != nil {
			logging.Logger.Error(err.Error())
		}
		buf.Write(jsonary)
	} else if p.MimeType == multipartData {
		// bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(&buf)
		// bodyWriter.CreateFormFile(p.Params[0], p.Params[0].Filename)

		if p.Params[0].Boundary != "" {
			bodyWriter.SetBoundary(p.Params[0].Boundary)
		}

		for _, param := range p.Params {
			if param.IsFile {
				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition",
					fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
						escapeQuotes(param.Name), escapeQuotes(param.Filename)))
				h.Set("Content-Type", param.ContentType)
				part, err := bodyWriter.CreatePart(h)
				if err != nil {
					logging.Logger.Error(err.Error())
				}
				// 写入文件数据到multipart，和读取本地文件方法的唯一区别
				_, err = part.Write([]byte(param.Value))
			} else {
				_ = bodyWriter.WriteField(param.Name, param.Value)
			}
		}
		bodyWriter.Close()
		// fmt.Println(buf.String())
	} else {
		for i, Param := range p.Params {
			buf.WriteString(Param.Name + "=" + Param.Value)
			if i != p.Len()-1 {
				buf.WriteString("&")
			}
		}
	}

	return buf.String()
}

func (p Variations) Set(key string, value string) error {
	for i, param := range p.Params {
		if param.Name == key {
			p.Params[i].Value = value
			return nil
		}
	}
	return fmt.Errorf("not found: %s", key)
}

/*
SetPayloadByIndex 根据索引设置payload

GET 返回 http://testphp.vulnweb.com/listproducts.php?artist=((,”"(,"","((

POST 返回 artist=')”,)'(())')(
*/
func (p *Variations) SetPayloadByIndex(index int, uri string, payload string, method string) string {
	// 对 payload 进行 url 编码
	payload = url.QueryEscape(payload)

	if strings.ToUpper(method) == "POST" {
		for idx, kv := range p.Params {
			// 判断是否为不可更改的参数名
			if util.SliceInCaseFold(kv.Name, util.ParamFilter) {
				continue
			}
			if idx == index {
				// 先改变参数，生成 payload，然后再将参数改回来，将现场还原
				p.Set(kv.Name, payload)
				str := p.Release()
				p.Set(kv.Name, p.OriginalParams[idx].Value)
				p.Release()
				return str
			}

		}
	} else if strings.ToUpper(method) == "GET" {
		u, err := url.Parse(uri)
		if err != nil {
			logging.Logger.Error(err.Error())
			return ""
		}
		v := u.Query()
		for idx, kv := range p.Params {
			// 判断是否为不可更改的参数名
			if util.SliceInCaseFold(kv.Name, util.ParamFilter) {
				continue
			}
			if idx == index {
				p.Set(kv.Name, payload)
				stv := p.Release()
				str := strings.Split(uri, "?")[0] + "?" + stv
				v.Set(kv.Name, kv.Value)

				p.Set(kv.Name, p.OriginalParams[idx].Value)
				p.Release()
				return str
			}
		}
	}

	return ""
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func Header(headers http.Header) string {
	var header string
	for key, value := range headers {
		header += fmt.Sprintf("%s: %s", key, strings.Join(value, "; "))
	}
	return header
}
