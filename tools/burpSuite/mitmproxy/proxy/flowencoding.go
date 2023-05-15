package proxy

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
	log "github.com/sirupsen/logrus"
)

var errEncodingNotSupport = errors.New("content-encoding not support")

var textContentTypes = []string{
	"text",
	"javascript",
	"json",
}

func (r *Response) IsTextContentType() bool {
	contentType := r.Header.Get("Content-Type")

	if contentType == "" {
		return false
	}
	for _, substr := range textContentTypes {
		if strings.Contains(contentType, substr) {
			return true
		}
	}
	return false
}

func (r *Response) DecodedBody() ([]byte, error) {
	if r.DecodedBodyStr != nil {
		return r.DecodedBodyStr, nil
	}

	if r.decodedErr != nil {
		return nil, r.decodedErr
	}

	if r.Body == nil {
		return nil, nil
	}

	if len(r.Body) == 0 {
		r.DecodedBodyStr = r.Body
		return r.DecodedBodyStr, nil
	}

	enc := r.Header.Get("Content-Encoding")
	if enc == "" || enc == "identity" {
		r.DecodedBodyStr = r.Body
		return r.DecodedBodyStr, nil
	}

	DecodedBodyStr, decodedErr := decode(enc, r.Body)
	if decodedErr != nil {
		r.decodedErr = decodedErr
		log.Error(r.decodedErr)
		return nil, decodedErr
	}

	r.DecodedBodyStr = DecodedBodyStr
	r.decoded = true
	return r.DecodedBodyStr, nil
}

func (r *Response) ReplaceToDecodedBody() {
	fmt.Println("999999\r\n", string(r.Body))
	fmt.Println("0000\r\n", string(r.DecodedBodyStr))
	body, err := r.DecodedBody()
	if err != nil || body == nil {
		return
	}

	r.Body = body

	fmt.Println("8888888\r\n", string(r.Body))
	r.Header.Del("Content-Encoding")
	r.Header.Set("Content-Length", strconv.Itoa(len(body)))
	r.Header.Del("Transfer-Encoding")
}

func decode(enc string, body []byte) ([]byte, error) {
	if enc == "gzip" {
		dreader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		buf := bytes.NewBuffer(make([]byte, 0))
		_, err = io.Copy(buf, dreader)
		if err != nil {
			return nil, err
		}
		err = dreader.Close()
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	} else if enc == "br" {
		dreader := brotli.NewReader(bytes.NewReader(body))
		buf := bytes.NewBuffer(make([]byte, 0))
		_, err := io.Copy(buf, dreader)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	} else if enc == "deflate" {
		dreader := flate.NewReader(bytes.NewReader(body))
		buf := bytes.NewBuffer(make([]byte, 0))
		_, err := io.Copy(buf, dreader)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	return nil, errEncodingNotSupport
}
