# Proxify 库逻辑分析

Proxify 是一个多功能的代理工具，设计为瑞士军刀，用于快速部署和流量操作。它主要基于 `martian/v3` 库进行 HTTP/HTTPS 的 MITM (Man-in-the-Middle) 代理，并结合了 `fastdialer` 进行网络连接，以及 `dsl` 引擎进行请求/响应的匹配和替换。它也支持 SOCKS5 代理和底层 TCP 套接字代理。

## 核心功能

-   **HTTP/HTTPS MITM 代理**: 能够拦截、查看、修改 HTTP 和 HTTPS 流量。
-   **SOCKS5 代理**: 支持作为 SOCKS5 服务器运行。
-   **TCP 套接字代理 (Socket Proxy)**: 可以在更低的层面代理 TCP 连接，并对非 HTTP 流量进行操作。
-   **流量转储 (Dumping)**: 可以将请求和响应的详细信息输出到文件 (JSONL, YAML, 或独立文件)。
-   **流量过滤与修改**:
    -   使用 DSL (Domain Specific Language) 表达式来匹配请求或响应。
    -   基于 DSL 表达式动态修改请求或响应的内容。
-   **上游代理**: 支持将流量转发到另一个 HTTP 或 SOCKS5 代理。
-   **自定义 DNS 解析**: 内置 DNS 服务器，可以自定义域名解析。
-   **回调机制**: 提供了在请求处理和响应处理阶段执行自定义逻辑的回调函数 (`OnRequestCallback`, `OnResponseCallback` for HTTP; `OnRequest`, `OnResponse` for SocketProxy)。

## 主要组件与结构

### 1. `Options` 结构体 (in `proxy.go`)

-   用于配置代理服务器的各种行为。关键字段包括：
    -   `ListenAddrHTTP`, `ListenAddrSocks5`: HTTP 和 SOCKS5 代理的监听地址。
    -   `OnRequestCallback`: `func(req *http.Request, ctx *martian.Context) error` - 在处理 HTTP 请求时调用的回调。
    -   `OnResponseCallback`: `func(resp *http.Response, ctx *martian.Context) error` - 在处理 HTTP 响应时调用的回调。
    -   `RequestDSL`, `ResponseDSL`: 用于匹配请求/响应的 DSL 表达式。
    -   `RequestMatchReplaceDSL`, `ResponseMatchReplaceDSL`: 用于修改请求/响应的 DSL 表达式。
    -   `DumpRequest`, `DumpResponse`, `OutputDirectory`, `OutputFile`, `OutputFormat`: 控制流量转储的选项。
    -   `UpstreamHTTPProxies`, `UpstreamSock5Proxies`: 上游代理设置。

### 2. `Proxy` 结构体 (in `proxy.go`)

-   代表一个 `proxify` 代理实例。
-   `httpProxy (*martian.Proxy)`: `martian` 库提供的核心 HTTP/HTTPS MITM 代理。
    -   `martian` 内部管理证书生成、TLS 握手等。
-   `logger (*logger.Logger)`: 用于记录流量和事件。
-   `Dialer (*fastdialer.Dialer)`: 用于建立出站连接。
-   **核心方法**:
    -   `NewProxy(options *Options)`: 创建并初始化一个新的 `Proxy` 实例。
    -   `Run()`: 启动 HTTP 和 SOCKS5 代理监听服务。
    -   `Stop()`: 停止代理服务。
    -   `ModifyRequest(req *http.Request) error`: `martian` 代理在收到客户端请求后调用的函数。这里会执行 DSL 匹配、`OnRequestCallback` 回调、DSL 修改和日志记录。
    -   `ModifyResponse(resp *http.Response) error`: `martian` 代理在收到上游服务器响应后调用的函数。类似地，执行 DSL 匹配、`OnResponseCallback` 回调、DSL 修改和日志记录。
    -   `setupHTTPProxy()`: 配置 `martian.Proxy` 实例，包括设置 MITM 配置、请求修饰器和响应修饰器。关键在于：
        -   `p.httpProxy.SetRequestModifier(p)`: 将 `Proxy` 自身设置为请求修饰器，这意味着 `Proxy.ModifyRequest` 会被调用。
        -   `p.httpProxy.SetResponseModifier(p)`: 将 `Proxy` 自身设置为响应修饰器，这意味着 `Proxy.ModifyResponse` 会被调用。

### 3. `SocketProxy` 结构体 (in `socket.go`)

-   用于处理非 HTTP 或底层 TCP 流量。
-   **`SocketProxyOptions`**: 配置 Socket 代理，关键字段：
    -   `OnRequest func([]byte) []byte`: 在收到客户端数据时调用，允许修改数据。
    -   `OnResponse func([]byte) []byte`: 在收到服务端数据时调用，允许修改数据。
    -   `RequestMatchReplaceDSL`, `ResponseMatchReplaceDSL`: 同样支持 DSL 修改原始字节流。
-   **核心方法**:
    -   `Run()`: 启动监听。
    -   `Proxy(conn net.Conn)`: 处理单个客户端连接，建立到目标服务器的连接。
    -   `pipe(src, dst io.ReadWriter)`: 在两个连接之间双向复制数据。在这个过程中，会调用 `OnRequest`/`OnResponse` 回调和 DSL 规则来处理/修改流经的数据。

### 4. DSL (Domain Specific Language)

-   `proxify` 使用 ProjectDiscovery 的 `dsl` 库 (`github.com/projectdiscovery/dsl`)。
-   允许用户编写简单的表达式来匹配和操作 HTTP 请求/响应的各个部分（如头部、正文、URL 等）或原始字节数据。
-   示例（来自 README）：
    -   匹配: `-request-dsl "contains(request,'firefox')"`
    -   替换: `-request-match-replace-dsl "replace(request,'firefox','chrome')"`

## HTTP/HTTPS 代理工作流程 (基于 `martian`)

1.  **启动**: `Proxy.Run()` 启动 `http.Server`，其 Handler 为 `Proxy.httpProxy`。
2.  **客户端连接**: 客户端发起 HTTP/HTTPS 请求到 `proxify` 监听的地址。
3.  **TLS MITM (对于 HTTPS)**:
    -   `martian` 拦截 TLS 连接。
    -   动态生成目标域名的伪造证书（使用预置或动态生成的 CA）。
    -   与客户端完成 TLS 握手，解密流量。
4.  **请求处理**:
    -   解密后的 `*http.Request` 进入 `Proxy.ModifyRequest` 方法。
    -   **DSL 匹配**: 根据 `options.RequestDSL` 判断请求是否满足条件。
    -   **`OnRequestCallback`**: 如果设置了此回调，则执行用户的自定义逻辑。**这是注入自定义处理的关键点。**
    -   **DSL 修改**: 根据 `options.RequestMatchReplaceDSL` 修改请求。
    -   **日志记录**: 请求被记录。
    -   请求被发送到目标服务器（或上游代理）。
5.  **响应处理**:
    -   从目标服务器收到 `*http.Response`。
    -   响应进入 `Proxy.ModifyResponse` 方法。
    -   **DSL 匹配**: 根据 `options.ResponseDSL` 判断响应是否满足条件。
    -   **`OnResponseCallback`**: 如果设置了此回调，则执行用户的自定义逻辑。**这是注入自定义处理的关键点。**
    -   **DSL 修改**: 根据 `options.ResponseMatchReplaceDSL` 修改响应。
    -   **日志记录**: 响应被记录。
    -   响应（可能已被修改和加密）发送回客户端。

## Socket 代理工作流程

1.  **启动**: `SocketProxy.Run()` 启动 TCP 监听。
2.  **客户端连接**: 客户端连接到 `SocketProxy` 监听的地址。
3.  **代理连接**: `SocketProxy` 连接到配置的远程地址。
4.  **数据管道**:
    -   `SocketConn.pipe()` 方法在客户端连接和服务器连接之间双向传输数据。
    -   **客户端到服务器**:
        -   读取客户端数据。
        -   应用 `RequestMatchReplaceDSL`。
        -   调用 `options.OnRequest([]byte) []byte` 回调。
        -   将（可能修改过的）数据发送到服务器。
    -   **服务器到客户端**:
        -   读取服务器数据。
        -   应用 `ResponseMatchReplaceDSL`。
        -   调用 `options.OnResponse([]byte) []byte` 回调。
        -   将（可能修改过的）数据发送到客户端。

## 使用 Proxify 重构 Burp Suite 功能的建议

本节将指导你如何使用 `proxify` 库替换现有基于 `go-mitmproxy` 的实现，以构建类似 Burp Suite 的核心功能：请求历史查看、流量拦截与修改。

### 1. 初始化和启动 Proxify

在你的 `main` 包或者应用启动的地方 (例如 `app.go` 中的 `Startup` 函数)，你需要：

-   **替换 `mitmproxy.Run()`**:
    -   创建一个 `proxify.Options` 实例。
    -   设置监听地址 (`ListenAddrHTTP`)，例如 `127.0.0.1:8888` (确保与前端配置一致)。
    -   **核心**: 实现并设置 `OnRequestCallback` 和 `OnResponseCallback` 函数。这些回调将包含你大部分的自定义逻辑。
    -   配置证书相关的选项，如 `CertCacheSize` 和 `Directory` (用于存储CA证书)。`proxify` 会处理 CA 证书的生成和管理。你需要确保 Wails 应用有权限在指定目录写入证书，并且用户能够方便地安装生成的 CA 证书到系统信任区或浏览器。
    -   调用 `proxify.NewProxy(options)` 创建代理实例。
    -   调用 `proxyInstance.Run()` 启动代理。

### 2. 实现回调函数

#### `onRequest` 回调
```go
func onRequest(req *http.Request, ctx *martian.Context) error {
    // 1. 过滤逻辑 (你可以根据需要实现自己的 Filter 函数)
    // if Filter(req.URL.Host) { return nil }
    // if req.Method == http.MethodConnect || req.Method == http.MethodOptions { return nil }

    reqDump, err := httputil.DumpRequestOut(req, true) 
    if err != nil {
        fmt.Printf("onRequest: Failed to dump request: %v
", err)
    }
    requestRawString := string(reqDump)
    flowID := ctx.ID()

    // 无论是否拦截，都先记录请求部分到缓存
    tempHistoryID := atomic.AddInt64(&historyItemIDGenerator, 1)
    tempEntry := &HTTPHistory{ // 使用你项目中的 HTTPHistory 结构
        Id: tempHistoryID,
        FlowID: flowID,
        Host: req.URL.Host,
        Method: req.Method,
        FullUrl: req.URL.String(),
        Path: req.URL.RequestURI(),
        RequestRaw: requestRawString,
        Timestamp: time.Now().Format(time.RFC3339),
    }
    TempHistoryCache.Store(flowID, tempEntry)
    // wailsapp.Event.Emit("HttpHistoryUpdateRequest", tempEntry) // 可选：立即通知前端请求已捕获

    if globalInterceptRequest {
        fmt.Printf("[Proxify] Intercepting Request ID: %s, URL: %s
", flowID, req.URL.String())
        interceptDataChan <- InterceptedData{
            ID:         flowID,
            Request:    cloneRequest(req), 
            MartianCtx: ctx,
            IsRequest:  true,
        }

        modifiedData, ok := <-forwardChan
        if !ok { // Channel closed
            return fmt.Errorf("forwardChan closed for request %s", flowID)
        }

        if modifiedData.ID != flowID {
            fmt.Printf("onRequest: Intercepted request ID mismatch. Expected %s, got %s. Forwarding original.
", flowID, modifiedData.ID)
        } else {
            switch modifiedData.Action {
            case "forward":
                if modifiedData.ModifiedBody != "" {
                    fmt.Printf("onRequest: Attempting to apply modified request for %s
", flowID)
                    newReq, parseErr := parseModifiedRequest(modifiedData.ModifiedBody, req.URL) 
                    if parseErr != nil {
                        fmt.Printf("onRequest: Failed to parse modified request %s: %v. Forwarding original.
", flowID, parseErr)
                    } else {
                        req.Method = newReq.Method
                        req.URL = newReq.URL
                        req.Header = newReq.Header
                        req.Body = newReq.Body
                        req.ContentLength = newReq.ContentLength
                        req.Host = newReq.Host // martian 可能依赖 req.Host
                        fmt.Printf("onRequest: Forwarding MODIFIED Request ID: %s
", flowID)
                        // 更新缓存中的原始请求串
                        updatedReqDump, _ := httputil.DumpRequestOut(req, true)
                        if cachedEntry, loaded := TempHistoryCache.Load(flowID); loaded {
                            cachedEntry.(*HTTPHistory).RequestRaw = string(updatedReqDump)
                        }
                    }
                } else {
                     fmt.Printf("onRequest: Forwarding ORIGINAL Request ID: %s
", flowID)
                }
            case "drop":
                fmt.Printf("onRequest: Dropping Request ID: %s
", flowID)
                TempHistoryCache.Delete(flowID) // 如果丢弃，也从缓存中移除
                return fmt.Errorf("request_dropped_by_user_%s", flowID) // martian 会关闭连接
            default:
                fmt.Printf("onRequest: Unknown action for request %s: %s. Forwarding original.
", flowID, modifiedData.Action)
            }
        }
    }
    return nil 
}
```

#### `onResponse` 回调
```go
func onResponse(resp *http.Response, ctx *martian.Context) error {
    req := resp.Request 
    // 1. 过滤逻辑
    // if Filter(req.URL.Host) { return nil }

    flowID := ctx.ID()

    if globalInterceptResponse {
        fmt.Printf("[Proxify] Intercepting Response ID: %s, URL: %s
", flowID, req.URL.String())
        
        // 确保响应体可以重复读取，因为后续处理和 DumpResponse 都需要读取
        var respBodyBytes []byte
        var errRead error
        if resp.Body != nil {
            respBodyBytes, errRead = io.ReadAll(resp.Body)
            if errRead != nil {
                fmt.Printf("onResponse: Error reading response body for dump: %v
", errRead)
                return errRead // 或者其他错误处理
            }
            resp.Body.Close() // 关闭原始 body
            resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes)) // 重新包装以便后续读取
        }

        interceptDataChan <- InterceptedData{
            ID:         flowID,
            Request:    cloneRequest(req), 
            Response:   cloneResponseWithBody(resp, respBodyBytes), // 发送包含已读body的副本
            MartianCtx: ctx,
            IsRequest:  false,
        }

        modifiedData, ok := <-forwardChan
        if !ok { // Channel closed
            return fmt.Errorf("forwardChan closed for response %s", flowID)
        }

        if modifiedData.ID != flowID {
            fmt.Printf("onResponse: Intercepted response ID mismatch. Expected %s, got %s. Forwarding original.
", flowID, modifiedData.ID)
        } else {
            switch modifiedData.Action {
            case "forward":
                if modifiedData.ModifiedBody != "" {
                    fmt.Printf("onResponse: Attempting to apply modified response for %s
", flowID)
                    newResp, parseErr := parseModifiedResponse(modifiedData.ModifiedBody, req) 
                    if parseErr != nil {
                        fmt.Printf("onResponse: Failed to parse modified response %s: %v. Forwarding original.
", flowID, parseErr)
                    } else {
                        resp.StatusCode = newResp.StatusCode
                        resp.Status = newResp.Status
                        resp.Header = newResp.Header
                        resp.Body = newResp.Body // newResp.Body 应该是可读的
                        resp.ContentLength = newResp.ContentLength
                        // ... 其他需要更新的字段 ...
                        fmt.Printf("onResponse: Forwarding MODIFIED Response ID: %s
", flowID)
                    }
                } else {
                    fmt.Printf("onResponse: Forwarding ORIGINAL Response ID: %s
", flowID)
                }
            case "drop":
                fmt.Printf("onResponse: Dropping Response ID: %s
", flowID)
                TempHistoryCache.Delete(flowID) // 如果丢弃，也从缓存中移除
                return fmt.Errorf("response_dropped_by_user_%s", flowID)
            default:
                fmt.Printf("onResponse: Unknown action for response %s: %s. Forwarding original.
", flowID, modifiedData.Action)
            }
        }
    }

    // 完成历史记录
    cachedEntry, loaded := TempHistoryCache.LoadAndDelete(flowID)
    if !loaded {
        fmt.Printf("onResponse: No request data found in cache for flow %s to complete history.
", flowID)
        // 即使没有请求部分，也可能需要记录一个不完整的历史条目
        // 或者，从 resp.Request 重新构建请求信息并记录
        // return nil // 或者根据策略决定是否报错
    }

    if entry, ok := cachedEntry.(*HTTPHistory); ok {
        // 再次确保响应体可读，因为 DumpResponse 会消耗它
        finalRespBodyBytes, errReadBody := io.ReadAll(resp.Body)
        if errReadBody != nil {
            fmt.Printf("onResponse: Error reading final response body for history: %v
", errReadBody)
        } else {
            resp.Body.Close()
            resp.Body = io.NopCloser(bytes.NewBuffer(finalRespBodyBytes)) // 重新包装
        }
        
        finalRespDump, _ := httputil.DumpResponse(resp, true) // Dump 最终返回给客户端的响应

        entry.Status = strconv.Itoa(resp.StatusCode)
        entry.Length = strconv.FormatInt(resp.ContentLength, 10) // ContentLength 可能不准确
        if len(finalRespBodyBytes) > 0 { // 使用实际读取到的 body 长度
             entry.Length = strconv.Itoa(len(finalRespBodyBytes))
        }
        entry.ContentType = resp.Header.Get("Content-Type")
        // entry.MIMEType = utils.GetMIMEType(entry.ContentType) // 你的工具函数
        // entry.Extension = utils.GetExtension(req.URL.Path)   // 你的工具函数
        // entry.Title = utils.GetTitle(string(finalRespBodyBytes)) // 你的工具函数
        entry.ResponseRaw = string(finalRespDump)
        entry.ResponseTimestamp = time.Now().Format(time.RFC3339)

        // wailsapp.Event.Emit("HttpHistoryUpdateFull", entry) // 通知前端完整条目
        fmt.Printf("History: Completed item ID %d for flow %s
", entry.Id, flowID)

        // TODO: 保存到数据库和你的持久化 HTTPBodyMap
        // db.SaveFullEntry(entry)
        // yourHTTPBodyMap.WriteMap(entry.Id, HTTPBody{...})
    }
    return nil
}
```

#### 辅助函数：克隆、解析、处理队列

```go
// cloneRequest 创建请求的副本
func cloneRequest(r *http.Request) *http.Request {
    r2 := r.Clone(r.Context()) 
    if r.Body != nil {
        bodyBytes, _ := io.ReadAll(r.Body)
        r.Body.Close() 
        r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
        r2.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
    }
    return r2
}

// cloneResponseWithBody 创建响应的副本，并使用预读的body
func cloneResponseWithBody(r *http.Response, bodyBytes []byte) *http.Response {
    r2 := *r
    r2.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 使用预读的body
    if r.Request != nil {
        r2.Request = cloneRequest(r.Request) 
    }
    return &r2
}


// parseModifiedRequest 将原始HTTP请求字符串解析为 *http.Request
func parseModifiedRequest(rawReq string, originalURL *url.URL) (*http.Request, error) {
    b := bufio.NewReader(strings.NewReader(rawReq))
    req, err := http.ReadRequest(b)
    if err != nil {
        return nil, fmt.Errorf("http.ReadRequest failed: %w", err)
    }
    // ReadRequest 通常不会设置 URL 的 Scheme 和 Host，需要从原始请求或上下文中获取
    req.URL.Scheme = originalURL.Scheme
    req.URL.Host = originalURL.Host
    if req.Host == "" { // ReadRequest 可能会将 Host 字段放在 Header 中
        req.Host = req.Header.Get("Host")
    }
     // 对于客户端请求，Host 字段应该等于 URL.Host
    if req.Host == "" {
        req.Host = originalURL.Host
    }

    // 如果是代理请求，RequestURI 应该被清除，让 http client 自己生成
    req.RequestURI = ""

    return req, nil
}

// parseModifiedResponse 将原始HTTP响应字符串解析为 *http.Response
func parseModifiedResponse(rawResp string, originalReq *http.Request) (*http.Response, error) {
    b := bufio.NewReader(strings.NewReader(rawResp))
    resp, err := http.ReadResponse(b, originalReq) // originalReq 用于 CONNECT 等请求
    if err != nil {
        return nil, fmt.Errorf("http.ReadResponse failed: %w", err)
    }
    return resp, nil
}

func processInterceptQueue() {
    for dataToIntercept := range interceptDataChan {
        eventPayload := make(map[string]interface{})
        eventPayload["id"] = dataToIntercept.ID
        var eventName string

        // 为了发送给前端，总是需要请求的dump
        reqForDisplay := cloneRequest(dataToIntercept.Request)
        reqDumpBytes, _ := httputil.DumpRequestOut(reqForDisplay, true)
        eventPayload["request"] = string(reqDumpBytes)

        if dataToIntercept.IsRequest {
            eventName = "ChYing.Event.InterceptRequestDisplay" // 与前端约定的事件名
            fmt.Printf("processInterceptQueue: Emitting %s for request ID %s
", eventName, dataToIntercept.ID)
        } else {
            respForDisplay := cloneResponseWithBody(dataToIntercept.Response, nil) // Body 在这里不需要，因为已经dump
            if dataToIntercept.Response.Body != nil { // 如果原始的有body
                 bodyBytesDump, _ := io.ReadAll(dataToIntercept.Response.Body) // 确保读取原始的
                 dataToIntercept.Response.Body.Close()
                 dataToIntercept.Response.Body = io.NopCloser(bytes.NewBuffer(bodyBytesDump))
                 respForDisplay.Body = io.NopCloser(bytes.NewBuffer(bodyBytesDump)) // 用它来dump
            }

            respDumpBytes, _ := httputil.DumpResponse(respForDisplay, true)
            eventPayload["response"] = string(respDumpBytes)
            eventName = "ChYing.Event.InterceptResponseDisplay" 
            fmt.Printf("processInterceptQueue: Emitting %s for response ID %s
", eventName, dataToIntercept.ID)
        }
        // wailsapp.Event.Emit(eventName, eventPayload) // 通过Wails app实例发送
    }
}
```

#### Wails 可调用函数 (在你的 App 结构体中)
```go
// // App struct (in app.go or similar)
// type App struct {
//     ctx analysis.Context
// }

// // SetWailsContext is called by Wails runtime
// func (a *App) SetWailsContext(ctx analysis.Context) {
//  a.ctx = ctx
// }

// // EmitEvent is a helper if you don't have a global wailsApp instance
// func (a *App) EmitEvent(eventName string, data ...interface{}) {
//  if a.ctx != nil {
//      runtime.EventsEmit(a.ctx, eventName, data...)
//  }
// }


// func (a *App) SetInterceptMode(isRequest bool, enable bool) {
//     if isRequest {
//         globalInterceptRequest = enable
//     } else {
//         globalInterceptResponse = enable
//     }
//     logging.Logger.Infof("Intercept mode: RequestEnabled=%v, ResponseEnabled=%v", globalInterceptRequest, globalInterceptResponse)
//     a.EmitEvent("ChYing.Event.InterceptStatusChanged", map[string]bool{
//         "requestActive": globalInterceptRequest,
//         "responseActive": globalInterceptResponse,
//     })
// }

// func (a *App) ForwardInterceptedData(id string, modifiedBody string, action string) {
//     // logging.Logger.Debugf("App.ForwardInterceptedData: ID=%s, Action=%s, BodyLen=%d", id, action, len(modifiedBody))
//     fmt.Printf("App.ForwardInterceptedData: ID=%s, Action=%s, BodyLen=%d
", id, action, len(modifiedBody))
//     select {
//     case forwardChan <- ModifiedData{
//         ID:           id,
//         ModifiedBody: modifiedBody,
//         Action:       action,
//     }:
//         // logging.Logger.Debugf("Forwarded data for ID %s to channel.", id)
//          fmt.Printf("Forwarded data for ID %s to channel.
", id)
//     default:
//         // logging.Logger.Warnf("forwardChan is full or not ready, could not send data for ID %s", id)
//         fmt.Printf("forwardChan is full or not ready, could not send data for ID %s
", id)
//         // 这里可能需要一些错误处理或通知前端
//     }
// }
```

### 3. 修改 `app.go` 和 `mitmproxy.go` (概述)

-   **移除 `go-mitmproxy`**: 大部分 `mitmproxy.go` (main package), `breakPoint.go`, `burpAddon.go` 的内容会被上述逻辑取代。
-   **保留和调整**:
    -   `app.go` 中的 `Startup` (现在会调用 `StartProxifyService`)。
    -   `App` 结构体的方法需要调整为调用新的拦截控制函数。
    -   数据存储逻辑 (`HTTPBodyMap`, 数据库操作) 需要被新的回调函数调用。
    -   `SendToRepeater`, `SendToIntruder` 等模块间交互的函数依然需要，但它们的数据来源是新的历史记录存储。
-   **ID管理**: `martian.Context.ID()` 作为 `FlowID`，原有的自增 `Id` 用于历史列表的显示和数据库主键。

### 4. 注意事项和挑战 (同前一部分，但在此重申关键点)

-   **HTTP报文解析/重建**: `parseModifiedRequest`/`Response` 的健壮性至关重要。
-   **并发安全**: 大量使用 `cloneRequest`/`cloneResponse`，确保channel和map的并发安全。
-   **错误处理与性能**: 回调函数中的错误和耗时操作会影响代理。
-   **证书**: `proxify` 会管理证书，引导用户安装 CA 证书 (`ca.crt`)。

这个重构方案提供了一个比较详细的起点。你需要根据你项目的具体结构和需求进行调整和细化。建议小步快跑，逐步替换和测试每个部分。 