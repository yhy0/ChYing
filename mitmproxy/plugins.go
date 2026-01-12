package mitmproxy

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

/**
   @author yhy
   @since 2025/7/2
   @desc 插件式请求/响应处理器管理
   @update 扩展支持只读和修改两种处理器模式
**/

// ProcessorMode 处理器模式枚举
type ProcessorMode int

const (
	// ReadOnly 只读模式 - 处理器只读取请求/响应，不修改原始数据
	// 会收到请求/响应的副本，其返回值会被忽略
	ReadOnly ProcessorMode = iota

	// Modifying 修改模式 - 处理器可以直接修改原始请求/响应
	// 如匹配替换功能应使用此模式
	Modifying
)

// RequestProcessorFunc 定义请求处理器函数类型
type RequestProcessorFunc func(req *http.Request) bool

// ResponseProcessorFunc 定义响应处理器函数类型
type ResponseProcessorFunc func(resp *http.Response) bool

// RequestProcessor 表示一个请求处理器
type RequestProcessor struct {
	Processor RequestProcessorFunc
	Mode      ProcessorMode
}

// ResponseProcessor 表示一个响应处理器
type ResponseProcessor struct {
	Processor ResponseProcessorFunc
	Mode      ProcessorMode
}

var (
	// 请求处理器列表
	requestProcessors     []RequestProcessor
	requestProcessorMutex sync.RWMutex

	// 响应处理器列表
	responseProcessors     []ResponseProcessor
	responseProcessorMutex sync.RWMutex
)

// RegisterRequestProcessor 注册一个请求处理器
// 默认为 ReadOnly 模式，不修改原始请求
func RegisterRequestProcessor(processor RequestProcessorFunc) {
	RegisterRequestProcessorWithMode(processor, ReadOnly)
}

// RegisterRequestProcessorWithMode 注册一个指定模式的请求处理器
func RegisterRequestProcessorWithMode(processor RequestProcessorFunc, mode ProcessorMode) {
	requestProcessorMutex.Lock()
	defer requestProcessorMutex.Unlock()
	requestProcessors = append(requestProcessors, RequestProcessor{
		Processor: processor,
		Mode:      mode,
	})
}

// RegisterResponseProcessor 注册一个响应处理器
// 默认为 ReadOnly 模式，不修改原始响应
func RegisterResponseProcessor(processor ResponseProcessorFunc) {
	RegisterResponseProcessorWithMode(processor, ReadOnly)
}

// RegisterResponseProcessorWithMode 注册一个指定模式的响应处理器
func RegisterResponseProcessorWithMode(processor ResponseProcessorFunc, mode ProcessorMode) {
	responseProcessorMutex.Lock()
	defer responseProcessorMutex.Unlock()
	responseProcessors = append(responseProcessors, ResponseProcessor{
		Processor: processor,
		Mode:      mode,
	})
}

// ProcessRequest 按注册顺序调用所有请求处理器
// 返回是否有任何处理器修改了请求
func ProcessRequest(req *http.Request) bool {
	requestProcessorMutex.RLock()
	defer requestProcessorMutex.RUnlock()

	modified := false

	// 先执行所有只读处理器（使用请求的副本）
	for _, processorEntry := range requestProcessors {
		if processorEntry.Mode == ReadOnly {
			reqCopy := cloneRequest(req)
			if reqCopy != nil {
				// 忽略返回值，因为是只读模式
				_ = processorEntry.Processor(reqCopy)
			}
		}
	}

	// 然后执行所有修改处理器（直接修改原始请求）
	for _, processorEntry := range requestProcessors {
		if processorEntry.Mode == Modifying {
			if processorEntry.Processor(req) {
				modified = true
			}
		}
	}

	return modified
}

// ProcessModifyingRequest 只执行修改类请求处理器
// 返回是否有任何处理器修改了请求
func ProcessModifyingRequest(req *http.Request) bool {
	requestProcessorMutex.RLock()
	defer requestProcessorMutex.RUnlock()

	modified := false

	// 只执行修改处理器（直接修改原始请求）
	for _, processorEntry := range requestProcessors {
		if processorEntry.Mode == Modifying {
			if processorEntry.Processor(req) {
				modified = true
			}
		}
	}

	return modified
}

// ProcessReadOnlyRequest 只执行只读类请求处理器，使用预读的body数据
func ProcessReadOnlyRequest(req *http.Request, bodyBytes []byte) {
	requestProcessorMutex.RLock()
	defer requestProcessorMutex.RUnlock()

	// 只执行只读处理器（使用请求的副本）
	for _, processorEntry := range requestProcessors {
		if processorEntry.Mode == ReadOnly {
			// 创建请求副本并设置Body
			reqCopy := req.Clone(req.Context())
			if bodyBytes != nil {
				reqCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			// 忽略返回值，因为是只读模式
			_ = processorEntry.Processor(reqCopy)
		}
	}
}

// ProcessResponse 按注册顺序调用所有响应处理器
// 返回是否有任何处理器修改了响应
func ProcessResponse(resp *http.Response) bool {
	responseProcessorMutex.RLock()
	defer responseProcessorMutex.RUnlock()

	modified := false

	// 先执行所有只读处理器（使用响应的副本）
	for _, processorEntry := range responseProcessors {
		if processorEntry.Mode == ReadOnly {
			respCopy := cloneResponseWithBody(resp, nil)
			if respCopy != nil {
				// 忽略返回值，因为是只读模式
				_ = processorEntry.Processor(respCopy)
			}
		}
	}

	// 然后执行所有修改处理器（直接修改原始响应）
	for _, processorEntry := range responseProcessors {
		if processorEntry.Mode == Modifying {
			if processorEntry.Processor(resp) {
				modified = true
			}
		}
	}

	return modified
}

// ProcessModifyingResponse 只执行修改类响应处理器
// 返回是否有任何处理器修改了响应
func ProcessModifyingResponse(resp *http.Response) bool {
	responseProcessorMutex.RLock()
	defer responseProcessorMutex.RUnlock()

	modified := false

	// 只执行修改处理器（直接修改原始响应）
	for _, processorEntry := range responseProcessors {
		if processorEntry.Mode == Modifying {
			if processorEntry.Processor(resp) {
				modified = true
			}
		}
	}

	return modified
}

// ProcessReadOnlyResponse 只执行只读类响应处理器，使用预读的body数据
func ProcessReadOnlyResponse(resp *http.Response, bodyBytes []byte) {
	responseProcessorMutex.RLock()
	defer responseProcessorMutex.RUnlock()

	// 只执行只读处理器（使用响应的副本）
	for _, processorEntry := range responseProcessors {
		if processorEntry.Mode == ReadOnly {
			// 创建响应副本并设置Body
			respCopy := *resp // 浅拷贝
			respCopy.Header = make(http.Header)
			for k, v := range resp.Header {
				respCopy.Header[k] = append([]string(nil), v...)
			}
			if bodyBytes != nil {
				respCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			if resp.Request != nil {
				respCopy.Request = resp.Request.Clone(resp.Request.Context())
			}
			// 忽略返回值，因为是只读模式
			_ = processorEntry.Processor(&respCopy)
		}
	}
}

// ClearAllProcessors 清除所有处理器
// 主要用于测试或重置系统
func ClearAllProcessors() {
	requestProcessorMutex.Lock()
	responseProcessorMutex.Lock()
	defer requestProcessorMutex.Unlock()
	defer responseProcessorMutex.Unlock()

	requestProcessors = nil
	responseProcessors = nil
}
