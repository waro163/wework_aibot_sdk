# 企业微信 AI Bot SDK for Go

[English](README.md) | [中文文档](README_zh.md)

功能完备且生产就绪的企业微信 AI Bot Go SDK，提供基于 WebSocket 的实时消息处理，支持自动重连、心跳检测和双模式消息传递。

## 特性

- 🔌 **WebSocket 客户端** - 完整的 WebSocket 客户端，支持连接生命周期管理
- 🔄 **自动重连** - 可配置的自动重连机制，支持指数退避
- 💓 **心跳机制** - 保持连接活跃的 ping/pong 心跳检测
- 🎯 **双模式消息传递** - 同时支持回调模式和 Channel 模式
- 🔒 **线程安全** - 并发安全实现，通过全面的竞争条件测试
- 📡 **事件驱动** - 支持消息、错误和重连事件的回调
- 🛡️ **生产就绪** - 广泛的测试覆盖，包括并发测试
- 🔐 **安全认证** - 内置企业微信服务器认证流程
- 📦 **文件下载** - HTTP 客户端支持从企业微信服务器下载文件
- 🔧 **可配置** - 精细控制超时、重试和时间间隔

## 安装

```bash
go get wework_aibot_sdk
```

## 快速开始

### 回调模式

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"

    sdk "wework_aibot_sdk"
)

func main() {
    cfg := &sdk.Config{
        BotId:             "your-bot-id",
        Secret:            "your-secret",
        HeartbeatInterval: 30,
        AutoReconnect:     true,
    }

    client, err := sdk.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // 设置消息处理器
    client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
        fmt.Printf("收到消息: %+v\n", msg)
        return nil
    })

    // 设置错误处理器
    client.SetErrorHandler(func(err error) {
        fmt.Printf("错误: %v\n", err)
    })

    // 设置重连处理器
    client.SetReconnectHandler(func() {
        fmt.Println("重连成功！")
    })

    // 启动客户端
    if err := client.Start(); err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    fmt.Println("客户端启动成功")

    // 等待中断信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
}
```

### Channel 模式

```go
package main

import (
    "fmt"
    "log"

    sdk "wework_aibot_sdk"
)

func main() {
    cfg := &sdk.Config{
        BotId:             "your-bot-id",
        Secret:            "your-secret",
        HeartbeatInterval: 30,
        AutoReconnect:     true,
    }

    client, err := sdk.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }

    if err := client.Start(); err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    // 从 channel 读取消息
    for msg := range client.Messages() {
        fmt.Printf("收到消息: %+v\n", msg)
    }
}
```

### 混合模式（回调 + Channel）

```go
// 在回调中处理关键消息
client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
    if msg.Body.MsgType == sdk.MSG_TYPE_TEXT {
        fmt.Printf("重要消息: %s\n", msg.Body.Text.Content)
    }
    return nil
})

// 异步处理所有消息
go func() {
    for msg := range client.Messages() {
        // 异步处理
        fmt.Printf("处理中: %+v\n", msg)
    }
}()
```

## 配置说明

```go
type Config struct {
    BotId                     string `json:"bot_id"`                      // 机器人 ID
    Secret                    string `json:"secret"`                      // 机器人密钥
    HeartbeatInterval         int    `json:"heartbeat_interval"`          // 心跳间隔（秒），默认 30
    AutoReconnect             bool   `json:"auto_reconnect"`              // 启用自动重连
    WsConnectionTimeout       int    `json:"ws_connection_timeout"`       // 连接超时（秒），默认 10
    WsConnectionMaxRetryTimes int    `json:"ws_connection_max_retry_times"` // 最大重试次数，默认 3
    WsConnectionRetryInterval int    `json:"ws_connection_retry_interval"`  // 重试间隔（秒），默认 2
    WsHandshakeTimeout        int    `json:"ws_handshake_timeout"`        // 握手超时（秒），默认 10
    SendQueueSize             int    `json:"send_queue_size"`             // 发送队列缓冲区大小，默认 100
    MsgChanSize               int    `json:"msg_chan_size"`               // 消息通道缓冲区大小，默认 100
}
```

## API 参考

### 客户端方法

#### `NewClient(cfg *Config) (*Client, error)`
创建新的 WebSocket 客户端实例。

#### `Start() error`
连接到服务器并进行认证。阻塞直到认证成功或失败。

#### `Stop() error`
优雅地关闭客户端，关闭所有 goroutine 和连接。

#### `Send(payload interface{}) error`
发送消息到服务器。接受任何可 JSON 序列化的 payload。

#### `State() ConnectionState`
获取当前连接状态（Disconnected、Connecting、Authenticating、Connected、Reconnecting）。

#### `Messages() <-chan CallbackPayload`

获取用于 channel 模式消费的消息通道。

#### `SetRawMessageHandler(handler func([]byte) error)`

设置原始消息字节回调（用于调试）。此处理器会接收所有消息的原始数据，包括认证、ping 和断连事件，在任何处理之前。

#### `SetMessageHandler(handler func(CallbackPayload) error)`

设置接收消息的回调函数。用户逻辑错误应由用户自己处理，此回调不应依赖 SDK 的错误处理器。

#### `SetErrorHandler(handler func(error))`

设置 SDK 内部错误的回调函数（连接错误、JSON 解析错误、消息通道满等），不包括用户业务逻辑错误。

#### `SetReconnectHandler(handler func())`

设置成功重连后的回调函数。

### 连接状态

```go
const (
    Disconnected     // 未连接
    Connecting       // 正在建立连接
    Authenticating   // 正在认证
    Connected        // 已连接并认证
    Reconnecting     // 重连中
)
```

## 消息类型

SDK 处理以下几种消息类型：

- **文本消息** - `MSG_TYPE_TEXT`
- **图片消息** - `MSG_TYPE_IMAGE`
- **事件消息** - `MSG_TYPE_EVENT`
- **语音消息** - `MSG_TYPE_AUDIO`
- **视频消息** - `MSG_TYPE_VIDEO`
- **文件消息** - `MSG_TYPE_FILE`
- **位置消息** - `MSG_TYPE_LOCATION`
- **链接消息** - `MSG_TYPE_LINK`
- **Markdown 消息** - `MSG_TYPE_MARKDOWN`

## 文件下载

### 下载并解密消息中的文件

企业微信会加密消息中的文件（图片、视频、文件等）。你需要使用提供的 `AesKey` 下载并解密：

```go
apiClient := sdk.NewApiClient(nil)

client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
    // 处理图片消息
    if msg.Body.MsgType == sdk.MSG_TYPE_IMAGE {
        if msg.Body.Image.Url != "" && msg.Body.Image.AesKey != "" {
            // 下载加密文件
            resp, err := apiClient.DownloadFileRaw(msg.Body.Image.Url)
            if err != nil {
                return fmt.Errorf("下载错误: %w", err)
            }

            // 解密文件数据
            imageData, err := sdk.DecryptFile(resp.FileData, msg.Body.Image.AesKey)
            if err != nil {
                return fmt.Errorf("解密错误: %w", err)
            }

            // 保存到文件
            f, err := os.Create(resp.FileName)
            if err != nil {
                return fmt.Errorf("创建文件错误: %w", err)
            }
            defer f.Close()

            _, err = f.Write(imageData)
            if err != nil {
                return fmt.Errorf("写入文件错误: %w", err)
            }

            fmt.Printf("已保存图片: %s\n", resp.FileName)
        }
    }

    // 类似地处理文件、视频等
    if msg.Body.MsgType == sdk.MSG_TYPE_FILE {
        if msg.Body.File.Url != "" && msg.Body.File.AesKey != "" {
            resp, err := apiClient.DownloadFileRaw(msg.Body.File.Url)
            if err != nil {
                return err
            }

            fileData, err := sdk.DecryptFile(resp.FileData, msg.Body.File.AesKey)
            if err != nil {
                return err
            }

            os.WriteFile(resp.FileName, fileData, 0644)
        }
    }

    return nil
})
```

### 下载普通文件

对于非加密文件或外部 URL：

```go
apiClient := sdk.NewApiClient(nil)

response, err := apiClient.DownloadFileRaw("https://example.com/file")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("文件名: %s\n", response.FileName)
fmt.Printf("大小: %d 字节\n", len(response.FileData))

// 保存到文件
err = os.WriteFile(response.FileName, response.FileData, 0644)
```

## 错误处理

SDK 区分可恢复和不可恢复的错误：

- **可恢复错误**：网络问题、临时断线 → 触发自动重连
- **不可恢复错误**：认证失败、JSON 序列化错误 → 停止重连

```go
var (
    ErrInvalidConfig      = errors.New("invalid config")          // 无效配置
    ErrMissingCredentials = errors.New("missing bot_id or secret") // 缺少凭证
    ErrNotConnected       = errors.New("not connected")            // 未连接
    ErrAlreadyStarted     = errors.New("client already started")   // 客户端已启动
    ErrAuthFailed         = errors.New("authentication failed")    // 认证失败
    ErrAuthTimeout        = errors.New("authentication timeout")   // 认证超时
    ErrUnrecoverable      = errors.New("unrecoverable error")      // 不可恢复错误
)
```

## 测试

```bash
# 运行所有测试
go test -v

# 使用竞争检测器运行
go test -race -v

# 仅运行并发测试
go test -race -v -run TestConcurrent

# 运行覆盖率测试
go test -cover
```

SDK 包含全面的测试：
- 所有组件的单元测试
- 并发测试（竞争条件检测）
- 死锁检测测试
- 连接生命周期测试

## 架构设计

客户端使用多 goroutine 架构和双 context 生命周期管理：

- **readLoop**: 从 WebSocket 接收消息
- **writeLoop**: 发送队列中的消息到 WebSocket
- **heartbeatLoop**: 发送周期性 ping 消息
- **reconnectionLoop**: 处理自动重连

线程安全通过以下机制实现：
- `connMu`: 保护连接和状态
- `callbackMu`: 保护回调函数
- `ctx` + `workerCtx`: 双 context 用于优雅关闭

## 使用示例

查看 [examples](examples/) 目录获取完整的工作示例：

- [client_example.go](examples/client_example.go) - 回调模式、Channel 模式和混合模式

## 高级用法

### 发送自定义消息

```go
// 发送文本消息
textPayload := map[string]interface{}{
    "cmd": "aibot_msg_send",
    "headers": map[string]string{
        "req_id": "custom-req-id-123",
    },
    "body": map[string]interface{}{
        "msg_type": "text",
        "text": map[string]string{
            "content": "Hello, WeWork!",
        },
    },
}

err := client.Send(textPayload)
if err != nil {
    log.Printf("发送失败: %v", err)
}
```

### 处理特定消息类型

```go
client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
    switch msg.Body.MsgType {
    case sdk.MSG_TYPE_TEXT:
        // 处理文本消息
        fmt.Printf("文本: %s\n", msg.Body.Text.Content)

    case sdk.MSG_TYPE_IMAGE:
        // 处理图片消息
        fmt.Printf("图片 URL: %s\n", msg.Body.Image.MediaId)

    case sdk.MSG_TYPE_FILE:
        // 下载文件
        apiClient := sdk.NewApiClient(nil)
        file, err := apiClient.DownloadFileRaw(msg.Body.File.FileUrl)
        if err != nil {
            return err
        }
        // 保存文件
        os.WriteFile(file.FileName, file.FileData, 0644)

    case sdk.MSG_TYPE_EVENT:
        // 处理事件
        fmt.Printf("事件类型: %s\n", msg.Body.Event.EventType)
    }
    return nil
})
```

### 优雅关闭

```go
client, _ := sdk.NewClient(cfg)
client.Start()

// 监听系统信号
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

<-sigChan
fmt.Println("正在关闭客户端...")

// 优雅关闭
if err := client.Stop(); err != nil {
    log.Printf("关闭错误: %v", err)
}

fmt.Println("客户端已关闭")
```

### 监控连接状态

```go
go func() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        state := client.State()
        fmt.Printf("当前状态: %s\n", state.String())
    }
}()
```

## 最佳实践

### 1. 错误处理

始终设置错误处理器来捕获运行时错误：

```go
client.SetErrorHandler(func(err error) {
    log.Printf("[错误] %v", err)
    // 可以在这里添加告警、日志等
})
```

### 2. 重连通知

使用重连处理器来处理重连后的逻辑：

```go
client.SetReconnectHandler(func() {
    log.Println("重连成功，重新订阅...")
    // 重新订阅、同步状态等
})
```

### 3. 消息处理超时

避免在消息处理器中执行长时间操作：

```go
client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
    // 快速处理或异步处理
    go processMessageAsync(msg)
    return nil
})
```

### 4. 调试和日志记录

使用原始消息处理器查看所有消息的原始数据，包括认证、ping、断连等内部消息：

```go
// 用于调试：记录所有原始消息
client.SetRawMessageHandler(func(data []byte) error {
    log.Printf("[原始消息] %s", string(data))
    return nil
})

// 更完整的示例：区分不同类型的消息
client.SetRawMessageHandler(func(data []byte) error {
    var msg map[string]interface{}
    if err := json.Unmarshal(data, &msg); err != nil {
        log.Printf("[原始消息-无法解析] %s", string(data))
        return nil
    }

    // 根据消息类型记录
    if cmd, ok := msg["cmd"].(string); ok {
        switch cmd {
        case "aibot_subscribe":
            log.Printf("[认证响应] %s", string(data))
        case "ping":
            log.Printf("[心跳响应] %s", string(data))
        case "aibot_event_callback":
            // 检查是否是断连事件
            if body, ok := msg["body"].(map[string]interface{}); ok {
                if event, ok := body["event"].(map[string]interface{}); ok {
                    if eventType, ok := event["eventtype"].(string); ok && eventType == "disconnected_event" {
                        log.Printf("[断连事件] %s", string(data))
                    }
                }
            }
        default:
            log.Printf("[业务消息] cmd=%s", cmd)
        }
    }

    return nil
})
```

**注意**：

- `SetRawMessageHandler` 会接收**所有**消息，包括内部消息
- 仅用于调试，生产环境建议关闭或只记录关键信息
- 原始消息处理器的错误不会触发 `SetErrorHandler`，需要自行处理

### 5. Channel 缓冲区

注意 channel 缓冲区大小（默认 100），如果消息处理慢可能会丢失消息：

```go
// 使用单独的 goroutine 处理
go func() {
    for msg := range client.Messages() {
        // 处理消息
        handleMessage(msg)
    }
}()
```

### 6. 优雅关闭

确保在程序退出时调用 `Stop()` 方法：

```go
defer client.Stop()
```

## 性能优化

### 并发安全

SDK 使用细粒度的锁来保证线程安全：
- `connMu` 仅在读写连接和状态时持有
- `callbackMu` 仅在读写回调函数时持有
- 用户回调在锁外执行，避免阻塞

### 内存管理

- Channel 使用缓冲区（100）避免阻塞
- goroutine 通过 context 确定性退出，避免泄漏
- 连接关闭时清理所有资源

### 消息队列

发送队列和消息通道都使用带缓冲的 channel，默认大小为 100。可以通过配置调整：

```go
cfg := &sdk.Config{
    BotId:         "your-bot-id",
    Secret:        "your-secret",
    SendQueueSize: 500,  // 增加发送队列大小到 500
    MsgChanSize:   500,  // 增加消息通道大小到 500
}
```

**何时需要增加缓冲区大小？**

- **SendQueueSize**: 如果频繁发送消息，且发送速度大于网络传输速度
- **MsgChanSize**: 如果接收消息速度大于处理速度，防止消息丢失

## 故障排查

### 1. 连接失败

```
failed to connect after N retries
```

**解决方法**：
- 检查网络连接
- 增加 `WsConnectionTimeout`
- 增加 `WsConnectionMaxRetryTimes`

### 2. 认证超时

```
authentication timeout
```

**解决方法**：
- 验证 `BotId` 和 `Secret` 是否正确
- 增加 `WsHandshakeTimeout`
- 检查企业微信后台配置

### 3. 消息丢失

**原因**：Channel 缓冲区满

**解决方法**：
- 加快消息处理速度
- 使用异步处理
- 考虑使用持久化队列

### 4. 内存泄漏

**检查**：
```bash
# 使用 pprof 检查
go test -memprofile=mem.out
go tool pprof mem.out
```

**常见原因**：
- 未调用 `Stop()` 关闭客户端
- goroutine 未正确退出

## 依赖要求

- Go 1.23.12 或更高版本
- 依赖库：
  - `github.com/gorilla/websocket v1.5.3`
  - `github.com/google/uuid v1.6.0`

## 许可证

MIT

## 贡献

欢迎贡献！请确保：
- 所有测试通过（`go test -race -v`）
- 代码格式正确（`go fmt`）
- 无竞争条件检测警告
- 新功能包含测试

## 参考

基于企业微信 AI Bot 官方 Node.js SDK: https://github.com/WecomTeam/aibot-node-sdk

## 常见问题 (FAQ)

### Q1: 如何同时使用回调和 Channel 模式？

A: 直接同时设置即可，SDK 会将消息分发到两个地方：

```go
client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
    // 回调处理
    return nil
})

go func() {
    for msg := range client.Messages() {
        // Channel 处理
    }
}()
```

### Q2: 自动重连会重新认证吗？

A: 是的，重连流程包括：
1. 关闭旧连接
2. 建立新连接
3. 重新发送认证
4. 认证成功后恢复消息处理

### Q3: 如何判断是否需要重连？

A: SDK 会自动判断，无需手动处理。可以通过 `SetReconnectHandler` 获得重连通知。

### Q4: 支持多个客户端实例吗？

A: 每个客户端实例独立管理连接和状态，但是企微服务不支持，如果多个客户端实例同时连接，会触发服务器disconnected_event事件，客户端收到该事件后，会断开连接。

