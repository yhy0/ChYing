package mitmproxy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/zstd"
	"github.com/yhy0/logging"
)

/**
   @author yhy
   @since 2025/7/15
   @desc HTTP内容处理工具函数
   提供HTTP请求和响应内容的处理功能，包括压缩内容解码和编码。
**/

// 处理后的HTTP响应体
type ProcessedResponseBody struct {
	// 解码后的内容
	Content []byte
	// 原始编码类型，如 "gzip", "br", "deflate" 等，为空表示未压缩
	OriginalEncoding string
	// 内容类型
	ContentType string
	// 是否为文本内容
	IsText bool
	// 原始压缩数据，仅当OriginalEncoding不为空时有效
	OriginalCompressedData []byte
	// 内容是否已被修改
	IsModified bool
}

// ProcessResponseBody 处理HTTP响应体，返回解码后的内容和原始编码类型
// 如果内容已经是未压缩的，则直接返回原内容
func ProcessResponseBody(resp *http.Response) (*ProcessedResponseBody, error) {
	if resp == nil || resp.Body == nil {
		return &ProcessedResponseBody{
			Content:                []byte{},
			OriginalEncoding:       "",
			ContentType:            "",
			IsText:                 false,
			OriginalCompressedData: nil,
			IsModified:             false,
		}, nil
	}

	// 读取原始响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 重置响应体以便后续处理
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	// 如果没有内容，直接返回
	if len(body) == 0 {
		return &ProcessedResponseBody{
			Content:                body,
			OriginalEncoding:       "",
			ContentType:            resp.Header.Get("Content-Type"),
			IsText:                 HttpIsTextContent(resp.Header.Get("Content-Type")),
			OriginalCompressedData: nil,
			IsModified:             false,
		}, nil
	}

	// 获取内容类型
	contentType := resp.Header.Get("Content-Type")
	isText := HttpIsTextContent(contentType)

	// 获取内容编码
	contentEncoding := resp.Header.Get("Content-Encoding")

	// 如果内容没有压缩或者不是文本内容(如图片、视频等)，直接返回原始内容
	// 对于二进制内容，即使有压缩也不进行解压，避免不必要的处理
	if contentEncoding == "" || contentEncoding == "identity" || !isText {
		if !isText && contentEncoding != "" && contentEncoding != "identity" {
			logging.Logger.Infof("检测到压缩的二进制内容(%s)，Content-Encoding: %s，跳过解压处理", contentType, contentEncoding)
		}

		return &ProcessedResponseBody{
			Content:                body,
			OriginalEncoding:       contentEncoding, // 保留原始编码信息，即使对二进制内容不解压
			ContentType:            contentType,
			IsText:                 isText,
			OriginalCompressedData: nil, // 不需要额外存储原始数据，因为Content就是原始数据
			IsModified:             false,
		}, nil
	}

	// 保存原始压缩数据
	originalCompressed := make([]byte, len(body))
	copy(originalCompressed, body)

	// 解码压缩内容
	decodedBody, decodedErr := httpDecodeCompressedBody(contentEncoding, body)
	if decodedErr != nil {
		logging.Logger.Errorf("解压响应体失败: %v，将使用原始内容", decodedErr)
		return &ProcessedResponseBody{
			Content:                body,
			OriginalEncoding:       "",
			ContentType:            contentType,
			IsText:                 isText,
			OriginalCompressedData: nil,
			IsModified:             false,
		}, nil
	}

	return &ProcessedResponseBody{
		Content:                decodedBody,
		OriginalEncoding:       contentEncoding,
		ContentType:            contentType,
		IsText:                 isText,
		OriginalCompressedData: originalCompressed,
		IsModified:             false,
	}, nil
}

// RecompressResponseBody 根据原始编码重新压缩响应体
// 如果originalEncoding为空，则直接返回原内容
func RecompressResponseBody(content []byte, originalEncoding string) ([]byte, error) {
	if originalEncoding == "" || originalEncoding == "identity" {
		return content, nil
	}

	compressedBody, err := httpEncodeCompressedBody(originalEncoding, content)
	if err != nil {
		return nil, err
	}

	return compressedBody, nil
}

// UpdateResponseWithProcessedBody 使用处理后的内容更新HTTP响应
// 如果需要恢复压缩，会重新压缩内容并更新Content-Encoding头
// 如果restoreCompression为false，则使用未压缩内容并移除Content-Encoding头
func UpdateResponseWithProcessedBody(resp *http.Response, processed *ProcessedResponseBody, restoreCompression bool) error {
	if resp == nil || processed == nil {
		return fmt.Errorf("response or processed body is nil")
	}

	var finalBody []byte
	var err error

	// 如果不是文本内容（如图片、视频等），直接使用原始内容
	if !processed.IsText && processed.OriginalEncoding != "" {
		// 对于二进制内容，总是使用原始内容，保持原始编码
		finalBody = processed.Content // 对于二进制内容，Content已经是原始数据
		// 保留原始Content-Encoding头
		resp.Header.Set("Content-Encoding", processed.OriginalEncoding)
	} else if restoreCompression && processed.OriginalEncoding != "" {
		// 处理文本内容，可能需要重新压缩

		// 优化：如果内容未被修改且有原始压缩数据，直接使用原始压缩数据
		if !processed.IsModified && len(processed.OriginalCompressedData) > 0 {
			finalBody = processed.OriginalCompressedData

		} else {
			// 内容已修改，需要重新压缩
			finalBody, err = RecompressResponseBody(processed.Content, processed.OriginalEncoding)
			if err != nil {
				logging.Logger.Warnf("重新压缩内容失败: %v，将使用未压缩内容", err)
				finalBody = processed.Content
				resp.Header.Del("Content-Encoding")
			} else {
				// 恢复Content-Encoding头
				resp.Header.Set("Content-Encoding", processed.OriginalEncoding)
			}
		}
	} else {
		// 使用未压缩内容
		finalBody = processed.Content
		// 移除Content-Encoding头
		resp.Header.Del("Content-Encoding")
	}

	// 更新响应体和Content-Length
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(finalBody))
	resp.ContentLength = int64(len(finalBody))
	resp.Header.Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))

	return nil
}

// SetContentModified 标记处理后的响应体内容已被修改
// 插件和匹配替换规则在修改内容后应调用此函数
func SetContentModified(processed *ProcessedResponseBody) {
	if processed != nil {
		processed.IsModified = true
	}
}

// GetDecodedResponseBody 获取解码后的响应体内容，不修改原始响应
// 主要用于前端显示和存储，返回解压后的可读内容
func GetDecodedResponseBody(resp *http.Response) ([]byte, error) {
	processed, err := ProcessResponseBody(resp)
	if err != nil {
		return nil, err
	}
	return processed.Content, nil
}

// GetResponseDumpWithDecodedBody 返回完整的响应dump，但使用解码后的响应体
// 用于存储到HTTPBodyMap以供前端显示
func GetResponseDumpWithDecodedBody(resp *http.Response) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("response is nil")
	}

	// 处理响应体
	processed, err := ProcessResponseBody(resp)
	if err != nil {
		return "", err
	}

	// 创建一个临时响应副本
	respCopy := *resp
	respCopy.Body = io.NopCloser(bytes.NewBuffer(processed.Content))
	respCopy.ContentLength = int64(len(processed.Content))
	respCopy.Header = resp.Header.Clone()
	respCopy.Header.Del("Content-Encoding")
	respCopy.Header.Set("Content-Length", strconv.FormatInt(respCopy.ContentLength, 10))

	// 添加一个注释头，标记这是解码后的内容
	if processed.OriginalEncoding != "" {
		respCopy.Header.Set("X-ChYing-Original-Encoding", processed.OriginalEncoding)
	}

	// Dump响应
	dumpBytes, err := httputil.DumpResponse(&respCopy, true)
	if err != nil {
		return "", err
	}

	return string(dumpBytes), nil
}

// 从这里开始是工具函数，从matchreplace.go提取并重命名以避免冲突

// httpDecodeCompressedBody 根据编码类型解压响应体
func httpDecodeCompressedBody(encoding string, body []byte) ([]byte, error) {
	var err error
	var decodedBody []byte

	// 处理多重编码情况，如 "gzip, deflate"
	encodings := strings.Split(encoding, ",")
	for _, enc := range encodings {
		enc = strings.TrimSpace(enc)

		// 选择解码方法
		switch enc {
		case "gzip":
			decodedBody, err = httpDecodeGzip(body)
		case "br":
			decodedBody, err = httpDecodeBrotli(body)
		case "deflate":
			decodedBody, err = httpDecodeDeflate(body)
		case "zstd":
			decodedBody, err = httpDecodeZstd(body)
		default:
			return nil, fmt.Errorf("不支持的编码类型: %s", enc)
		}

		if err != nil {
			return nil, fmt.Errorf("解码 %s 编码失败: %v", enc, err)
		}

		// 如果有多重编码，更新用于下一次解码的数据
		body = decodedBody
	}

	return decodedBody, nil
}

// httpDecodeGzip 解压 gzip 编码的数据
func httpDecodeGzip(body []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buf := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// httpDecodeDeflate 解压 deflate 编码的数据
func httpDecodeDeflate(body []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(body))
	defer reader.Close()

	buf := bytes.NewBuffer(make([]byte, 0))
	_, err := io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// httpDecodeBrotli 解压 brotli 编码的数据
func httpDecodeBrotli(body []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(body))
	buf := bytes.NewBuffer(make([]byte, 0))
	_, err := io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// httpDecodeZstd 解压 zstd 编码的数据
func httpDecodeZstd(body []byte) ([]byte, error) {
	reader, err := zstd.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buf := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// httpEncodeCompressedBody 根据编码类型压缩响应体
func httpEncodeCompressedBody(encoding string, body []byte) ([]byte, error) {
	encoding = strings.TrimSpace(encoding)

	// 简化处理：如果有多重编码，只使用第一种
	if strings.Contains(encoding, ",") {
		encoding = strings.Split(encoding, ",")[0]
		encoding = strings.TrimSpace(encoding)
	}

	switch encoding {
	case "gzip":
		return httpEncodeGzip(body)
	case "br":
		return httpEncodeBrotli(body)
	case "deflate":
		return httpEncodeDeflate(body)
	case "zstd":
		return httpEncodeZstd(body)
	default:
		return nil, fmt.Errorf("不支持的编码类型: %s", encoding)
	}
}

// httpEncodeGzip 压缩为 gzip 格式
func httpEncodeGzip(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err := gzipWriter.Write(body)
	if err != nil {
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// httpEncodeDeflate 压缩为 deflate 格式
func httpEncodeDeflate(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	deflateWriter, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}

	_, err = deflateWriter.Write(body)
	if err != nil {
		return nil, err
	}

	err = deflateWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// httpEncodeBrotli 压缩为 brotli 格式
func httpEncodeBrotli(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	brotliWriter := brotli.NewWriter(&buf)

	_, err := brotliWriter.Write(body)
	if err != nil {
		return nil, err
	}

	err = brotliWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// httpEncodeZstd 压缩为 zstd 格式
func httpEncodeZstd(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	zstdWriter, err := zstd.NewWriter(&buf)
	if err != nil {
		return nil, err
	}

	_, err = zstdWriter.Write(body)
	if err != nil {
		return nil, err
	}

	err = zstdWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// HttpIsTextContent 判断内容类型是否可能是文本
func HttpIsTextContent(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/xml") ||
		strings.Contains(contentType, "application/javascript") ||
		strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "+json") ||
		strings.Contains(contentType, "+xml") ||
		(contentType == "" || contentType == "application/octet-stream") // 未知类型也尝试处理
}
