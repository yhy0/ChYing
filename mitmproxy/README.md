# ChYing-Inside 代理系统 (mitmproxy)

本目录包含了ChYing-Inside的HTTP代理系统核心代码，实现了HTTP请求拦截、修改和分析功能。其核心基于 `lib/proxify` (martian) 构建，并通过插件系统提供高度的可扩展性。

## 核心回调与插件系统

代理的核心逻辑围绕 `proxify.go` 中的 `onRequestCallback` 和 `onResponseCallback` 函数展开。这两个回调函数会调用 `plugins.go` 中定义的插件处理器链。

### 插件处理器模式 (`plugins.go`)

插件系统支持两种处理器模式 (`ProcessorMode`)，定义在 `plugins.go`：

1.  **只读模式 (ReadOnly)**
    *   处理器接收请求/响应的**深拷贝副本**。对原始数据无直接影响。
    *   适用于：日志记录、安全扫描、流量分析等不需要修改原始数据的场景。
    *   返回值会被忽略。
    *   **执行顺序**：所有 `ReadOnly` 处理器在相应的 `Modifying` 处理器之前执行。

2.  **修改模式 (Modifying)**
    *   处理器直接接收**原始的请求/响应对象**，并可以对其进行修改。
    *   适用于：内容替换、请求/响应修改、安全过滤等需要改变HTTP数据的场景。
    *   返回 `true` 表示数据已被修改，`false` 表示未修改。
    *   **执行顺序**：所有 `Modifying` 处理器在相应的 `ReadOnly` 处理器之后执行。

### 请求处理流程 (`onRequestCallback` -> `ProcessRequest`)

1.  HTTP请求到达，进入 `onRequestCallback`。
2.  调用 `ProcessRequest` 处理器链：
    *   **首先**，所有注册的 `ReadOnly` 请求处理器按注册顺序执行，它们操作的是请求的副本。
        *   例如，`authcheck.go` 的 `captureRequestForAuthTesting`。
    *   **然后**，所有注册的 `Modifying` 请求处理器按注册顺序执行，它们操作的是原始请求对象。
        *   例如，`matchreplace.go` 的 `MatchReplaceRequestProcessor`。
3.  如果任何 `Modifying` 处理器修改了请求，原始请求对象将被更新。
4.  （可选）用户通过UI进行的实时拦截和修改会在此流程中或之后应用。

### 响应处理流程 (`onResponseCallback` -> `ProcessResponse`)

1.  HTTP响应到达，进入 `onResponseCallback`。
2.  首先，调用 `httputil.ProcessResponseBody` 对响应进行初步处理（如解压），得到 `processedInfo`。
3.  （可选）用户通过UI进行的实时拦截和修改首先应用。
4.  调用 `ProcessResponse` 处理器链：
    *   **首先**，所有注册的 `ReadOnly` 响应处理器按注册顺序执行，它们操作的是当前响应状态的副本。
        *   例如，`passiveScanPlugin.go` 的 `passiveScanResponseProcessor`。
    *   **然后**，所有注册的 `Modifying` 响应处理器按注册顺序执行，它们操作的是原始响应对象。
        *   例如，`matchreplace.go` 的 `MatchReplaceResponseProcessor`。
5.  如果任何 `Modifying` 处理器修改了响应，`onResponseCallback` 会再次调用 `httputil.ProcessResponseBody` 来获取基于最新修改的 `processedInfo`。
6.  最后，`onResponseCallback` 使用最终的 `processedInfo` 和 `httputil.UpdateResponseWithProcessedBody` 来准备将要发送给客户端的响应（包括必要的重压缩和头部更新）。
7.  用于存储到 `HTTPBodyMap` 的 `ResponseRaw` 是基于最终的、解码后的响应内容生成的。

## HTTP内容处理工具 (`httputil.go`)

为了确保所有插件都能在可理解的数据上操作，并且最终用户能看到正确的响应内容，`httputil.go` 提供了关键的内容处理能力：

*   **自动解压缩/压缩**：支持 gzip, brotli, deflate, zstd。`ProcessResponseBody` 会尝试解码文本内容，而 `UpdateResponseWithProcessedBody` 会根据原始编码和内容是否被修改来决定是否重压缩。
*   **文本内容识别**：`HttpIsTextContent` 用于判断内容类型是否适合进行解压和文本处理。
*   **状态追踪**：`ProcessedResponseBody` 结构体用于在处理流程中传递解码后的内容、原始编码、是否被修改等状态。
*   **为存储生成可读 Dump**：`dumpResponseFromProcessedInfo` (在 `proxify.go` 中) 或 `GetResponseDumpWithDecodedBody` (在 `httputil.go` 中) 用于为 `HTTPBodyMap` 生成包含解压后 Body 的响应字符串。

## 主要插件及其行为

### 1. 内容匹配与替换 (`matchreplace.go`)

*   **模式**: `Modifying` (请求和响应)。
*   **执行时机**: 在相应路径的 `ReadOnly` 插件之后。
*   **功能**:
    *   允许用户定义规则（支持正则表达式）来匹配和替换请求/响应的头部、正文、首行、URL参数等。
    *   对于响应体，它会利用 `httputil.ProcessResponseBody` 获取解码内容，在其上进行替换，然后调用 `httputil.UpdateResponseWithProcessedBody` 来更新原始响应（包括重压缩）。
    *   是改变实际代理流量内容的主要插件。

### 2. 越权检测 (`authcheck.go`)

*   **模式**: `ReadOnly` (请求)。
*   **执行时机**: 在请求路径的 `Modifying` 插件（如 `matchreplace.go`）**之前**。
*   **功能**:
    *   捕获**原始HTTP请求**的副本。
    *   **异步地**（通过 goroutine）基于这些副本创建新的测试请求（例如，修改/删除认证相关的头部）。
    *   发送这些测试请求并记录响应，用于检测潜在的越权漏洞。
    *   其操作不直接影响流经代理的原始请求/响应对。

### 3. 被动扫描 (`passiveScanPlugin.go`)

*   **模式**: `ReadOnly` (响应)。
*   **执行时机**: 在响应路径的 `Modifying` 插件（如 `matchreplace.go`）**之前**。
*   **功能**:
    *   获取**当前响应状态**的副本（即，可能已被用户UI拦截修改，但尚未被 `matchreplace.go` 修改）。
    *   **异步地**（通过 goroutine）对这个副本进行处理：
        *   调用 `GetDecodedResponseBody` 获取解码后的响应内容。
        *   基于解码后的内容执行被动安全扫描规则。
    *   其操作不直接影响流经代理的原始响应。

## 注册处理器示例

在 `InitProxifyState` 或相应插件的初始化函数中：

```go
// 注册只读请求处理器 (authcheck)
RegisterRequestProcessorWithMode(captureRequestForAuthTesting, ReadOnly)

// 注册修改模式请求处理器 (matchreplace)
RegisterRequestProcessorWithMode(MatchReplaceRequestProcessor, Modifying)

// 注册只读响应处理器 (passiveScan)
RegisterResponseProcessorWithMode(passiveScanResponseProcessor, ReadOnly)

// 注册修改模式响应处理器 (matchreplace)
RegisterResponseProcessorWithMode(MatchReplaceResponseProcessor, Modifying)
```

## 数据存储与事件通知

*   **HTTP历史 (`HTTPHistory`)**: 存储每个请求/响应的基本元数据，通过 `EventData` 发送给前端列表。
*   **原始报文 (`HTTPBodyMap`)**: 键为 `HTTPHistory.Id`，值为 `HTTPBody` 结构（包含 `RequestRaw` 和 `ResponseRaw`）。`ResponseRaw` 存储的是**解码后、可读的**响应内容，由 `dumpResponseFromProcessedInfo` 生成。
*   **事件通道 (`EventData`)**: 用于将新的 `HTTPHistory` 事件以及其他代理事件发送到应用核心，供前端消费。

## 性能与异步处理

*   `authcheck.go` 和 `passiveScanPlugin.go` 的核心分析逻辑已改为异步执行，以减少对主代理线程的阻塞，提高响应速度。
*   响应体的处理（特别是压缩/解压缩）是性能敏感点，`onResponseCallback` 中的逻辑经过优化，试图在保证功能正确的前提下减少不必要的重复操作。

## 调试与排错

*   关注 `logging.Logger` 输出的日志，特别是关于内容处理、插件执行和错误的记录。
*   检查 `Content-Length` 和 `Content-Encoding` 头部是否与最终发送的响应体一致。
*   如果出现乱码或内容错误，追踪 `httputil.go` 中的处理函数以及 `onResponseCallback` 中 `processedInfo` 的状态变化。 