package httpx

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/projectdiscovery/rawhttp/client"
	"github.com/projectdiscovery/stringsutil"
	errorutil "github.com/projectdiscovery/utils/errors"
	urlutil "github.com/projectdiscovery/utils/url"
	"github.com/yhy0/logging"
	"io"
	"strings"
)

/**
  @author: yhy
  @since: 2023/4/25
  @desc: 提取自 nuclei https://github.com/projectdiscovery/nuclei/tree/main/v2/pkg/protocols/http/raw/raw.go
**/

// RawRequest defines a basic HTTP raw request
type RawRequest struct {
	FullURL        string
	Method         string
	Path           string
	Data           string
	Headers        map[string]string
	UnsafeHeaders  client.Headers
	UnsafeRawBytes []byte
}

// Raw 通过 raw 格式直接直接请求 todo 现在修改 host 对应的值，并不会改变
func Raw(request string, target string) (*Response, error) {
	inputURL, err := urlutil.ParseURL(target, true)

	if err != nil {
		logging.Logger.Errorln(target, err)
		return nil, err
	}

	rawRequestData, err := Parse(request, inputURL, true)

	if err != nil {
		logging.Logger.Errorln(err)
		return nil, err
	}

	return RequestRaw(rawRequestData.FullURL, rawRequestData.Method, rawRequestData.Data, false, rawRequestData.Headers)
}

// Parse parses the raw request as supplied by the user
func Parse(request string, inputURL *urlutil.URL, unsafe bool) (*RawRequest, error) {
	rawrequest, err := readRawRequest(request, unsafe)
	if err != nil {
		return nil, err
	}

	switch {
	// If path is empty do not tamper input url (see doc)
	// can be omitted but makes things clear
	case rawrequest.Path == "":
		rawrequest.Path = inputURL.GetRelativePath()

	// full url provided instead of rel path
	case strings.HasPrefix(rawrequest.Path, "http") && !unsafe:
		urlx, err := urlutil.ParseURL(rawrequest.Path, true)
		if err != nil {
			return nil, errorutil.NewWithErr(err).WithTag("raw").Msgf("failed to parse url %v from template", rawrequest.Path)
		}
		cloned := inputURL.Clone()
		parseErr := cloned.MergePath(urlx.GetRelativePath(), true)
		if parseErr != nil {
			return nil, errorutil.NewWithTag("raw", "could not automergepath for template path %v", urlx.GetRelativePath()).Wrap(parseErr)
		}
		rawrequest.Path = cloned.GetRelativePath()
	// If unsafe changes must be made in raw request string iteself
	case unsafe:
		prevPath := rawrequest.Path
		cloned := inputURL.Clone()
		unsafeRelativePath := ""
		if (cloned.Path == "" || cloned.Path == "/") && !strings.HasPrefix(prevPath, "/") {
			// Edgecase if raw unsafe request is
			// GET 1337?with=param HTTP/1.1
			if tmpurl, err := urlutil.ParseRelativePath(prevPath, true); err == nil && len(tmpurl.Params) > 0 {
				// if raw request contains parameters
				cloned.Params.Merge(tmpurl.Params)
				unsafeRelativePath = strings.TrimPrefix(tmpurl.Path, "/") + "?" + cloned.Params.Encode()
			} else {
				// if raw request does not contain param
				if len(cloned.Params) > 0 {
					unsafeRelativePath = prevPath + "?" + cloned.Params.Encode()
				} else {
					unsafeRelativePath = prevPath
				}
			}
		} else {
			err = cloned.MergePath(rawrequest.Path, true)
			if err != nil {
				return nil, errorutil.NewWithErr(err).WithTag("raw").Msgf("failed to automerge %v from unsafe template", rawrequest.Path)
			}
			unsafeRelativePath = cloned.GetRelativePath()
		}
		rawrequest.UnsafeRawBytes = bytes.Replace(rawrequest.UnsafeRawBytes, []byte(prevPath), []byte(unsafeRelativePath), 1)

	default:
		cloned := inputURL.Clone()
		parseErr := cloned.MergePath(rawrequest.Path, true)
		if parseErr != nil {
			return nil, errorutil.NewWithTag("raw", "could not automergepath for template path %v", rawrequest.Path).Wrap(parseErr)
		}
		rawrequest.Path = cloned.GetRelativePath()
	}

	if !unsafe {
		if _, ok := rawrequest.Headers["Host"]; !ok {
			rawrequest.Headers["Host"] = inputURL.Host
		}
	}
	rawrequest.FullURL = fmt.Sprintf("%s://%s%s", inputURL.Scheme, strings.TrimSpace(inputURL.Host), rawrequest.Path)

	return rawrequest, nil
}

// reads raw request line by line following convention
func readRawRequest(request string, unsafe bool) (*RawRequest, error) {
	rawRequest := &RawRequest{
		Headers: make(map[string]string),
	}

	// store body if it is unsafe request
	if unsafe {
		rawRequest.UnsafeRawBytes = []byte(request)
	}

	// parse raw request
	reader := bufio.NewReader(strings.NewReader(request))
read_line:
	s, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read request: %w", err)
	}
	// ignore all annotations
	if stringsutil.HasPrefixAny(s, "@") {
		goto read_line
	}

	parts := strings.Split(s, " ")
	if len(parts) > 0 {
		rawRequest.Method = parts[0]
		if len(parts) == 2 && strings.Contains(parts[1], "HTTP") {
			// When relative path is missing/ not specified it is considered that
			// request is meant to be untampered at path
			// Ex: GET HTTP/1.1
			parts = []string{parts[0], "", parts[1]}
		}
		if len(parts) < 3 && !unsafe {
			// missing a field
			return nil, fmt.Errorf("malformed request specified: %v", s)
		}

		// relative path
		rawRequest.Path = parts[1]
		// Note: raw request does not URL Encode if needed `+` should be used
		// this can be also be implemented
	}

	var multiPartRequest bool
	// Accepts all malformed headers
	var key, value string
	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if readErr != nil || line == "" {
			if readErr != io.EOF {
				break
			}
		}

		p := strings.SplitN(line, ":", 2)
		key = p[0]
		if len(p) > 1 {
			value = p[1]
		}
		if strings.Contains(key, "Content-Type") && strings.Contains(value, "multipart/") {
			multiPartRequest = true
		}

		// in case of unsafe requests multiple headers should be accepted
		// therefore use the full line as key
		_, found := rawRequest.Headers[key]
		if unsafe {
			rawRequest.UnsafeHeaders = append(rawRequest.UnsafeHeaders, client.Header{Key: line})
		}

		if unsafe && found {
			rawRequest.Headers[line] = ""
		} else {
			rawRequest.Headers[key] = strings.TrimSpace(value)
		}
		if readErr == io.EOF {
			break
		}
	}

	// Set the request body
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("could not read request body: %w", err)
	}
	rawRequest.Data = string(b)
	if !multiPartRequest {
		rawRequest.Data = strings.TrimSuffix(rawRequest.Data, "\r\n")
	}
	return rawRequest, nil

}
