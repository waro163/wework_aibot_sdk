# WeWork AI Bot SDK for Go

[English](README.md) | [中文文档](README_zh.md)

A robust and production-ready Go SDK for WeWork (WeCom) AI Bot, providing WebSocket-based real-time message handling with auto-reconnection, heartbeat, and dual message delivery modes.

## Features

- 🔌 **WebSocket Client** - Full-featured WebSocket client with connection lifecycle management
- 🔄 **Auto-Reconnection** - Configurable automatic reconnection with exponential backoff
- 💓 **Heartbeat Mechanism** - Keep-alive ping/pong to maintain connection health
- 🎯 **Dual Message Delivery** - Support both callback-based and channel-based message handling simultaneously
- 🔒 **Thread-Safe** - Concurrent-safe implementation with comprehensive race condition testing
- 📡 **Event-Driven** - Callbacks for messages, errors, and reconnection events
- 🛡️ **Production-Ready** - Extensive test coverage with concurrency testing
- 🔐 **Secure Authentication** - Built-in authentication flow with WeWork servers
- 📦 **File Download** - HTTP client for downloading files from WeWork servers
- 🔧 **Configurable** - Fine-grained control over timeouts, retries, and intervals

## Installation

```bash
go get github.com/waro163/wework_aibot_sdk
```

## Quick Start

### Callback Mode

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"

    sdk "github.com/waro163/wework_aibot_sdk"
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

    // Set message handler
    client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
        fmt.Printf("Received: %+v\n", msg)
        return nil
    })

    // Set error handler
    client.SetErrorHandler(func(err error) {
        fmt.Printf("Error: %v\n", err)
    })

    // Set reconnect handler
    client.SetReconnectHandler(func() {
        fmt.Println("Reconnected successfully!")
    })

    // Start client
    if err := client.Start(); err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    fmt.Println("Client started successfully")

    // Wait for interrupt
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
}
```

### Channel Mode

```go
package main

import (
    "fmt"
    "log"

    sdk "github.com/waro163/wework_aibot_sdk"
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

    // Read from channel
    for msg := range client.Messages() {
        fmt.Printf("Received: %+v\n", msg)
    }
}
```

### Hybrid Mode (Callback + Channel)

```go
client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
    // Handle critical messages immediately
    if msg.Body.MsgType == sdk.MSG_TYPE_TEXT {
        fmt.Printf("Critical: %s\n", msg.Body.Text.Content)
    }
    return nil
})

// Process all messages asynchronously
go func() {
    for msg := range client.Messages() {
        // Async processing
        fmt.Printf("Processing: %+v\n", msg)
    }
}()
```

## Configuration

```go
type Config struct {
    BotId                     string `json:"bot_id"`
    Secret                    string `json:"secret"`
    HeartbeatInterval         int    `json:"heartbeat_interval"`          // seconds, default 30
    AutoReconnect             bool   `json:"auto_reconnect"`              // enable auto-reconnect
    WsConnectionTimeout       int    `json:"ws_connection_timeout"`       // seconds, default 10
    WsConnectionMaxRetryTimes int    `json:"ws_connection_max_retry_times"` // default 3
    WsConnectionRetryInterval int    `json:"ws_connection_retry_interval"`  // seconds, default 2
    WsHandshakeTimeout        int    `json:"ws_handshake_timeout"`        // seconds, default 10
    SendQueueSize             int    `json:"send_queue_size"`             // send queue buffer size, default 100
    MsgChanSize               int    `json:"msg_chan_size"`               // message channel buffer size, default 100
}
```

## API Reference

### Client Methods

#### `NewClient(cfg *Config) (*Client, error)`
Create a new WebSocket client instance.

#### `Start() error`
Connect to server and authenticate. Blocks until authentication succeeds or fails.

#### `Stop() error`
Gracefully shutdown the client, closing all goroutines and connections.

#### `Send(payload interface{}) error`
Send a message to the server. Accepts any JSON-marshalable payload.

#### `State() ConnectionState`
Get current connection state (Disconnected, Connecting, Authenticating, Connected, Reconnecting).

#### `Messages() <-chan CallbackPayload`

Get the message channel for channel-based consumption.

#### `SetRawMessageHandler(handler func([]byte) error)`

Set callback for raw message bytes (for debugging). This handler receives all messages including auth, ping, and disconnect events before any processing.

#### `SetMessageHandler(handler func(CallbackPayload) error)`

Set callback for incoming messages. User-level errors should be handled by the user; this callback should not rely on SDK's error handler.

#### `SetErrorHandler(handler func(error))`

Set callback for SDK internal errors (connection errors, JSON parsing errors, channel full, etc.), not user logic errors.

#### `SetReconnectHandler(handler func())`

Set callback for successful reconnections.

### Connection States

```go
const (
    Disconnected     // Not connected
    Connecting       // Establishing connection
    Authenticating   // Sending authentication
    Connected        // Fully connected and authenticated
    Reconnecting     // Reconnecting after disconnection
)
```

## Message Types

The SDK handles several message types:

- **Text Messages** - `MSG_TYPE_TEXT`
- **Image Messages** - `MSG_TYPE_IMAGE`
- **Event Messages** - `MSG_TYPE_EVENT`
- **Audio Messages** - `MSG_TYPE_AUDIO`
- **Video Messages** - `MSG_TYPE_VIDEO`
- **File Messages** - `MSG_TYPE_FILE`
- **Location Messages** - `MSG_TYPE_LOCATION`
- **Link Messages** - `MSG_TYPE_LINK`
- **Markdown Messages** - `MSG_TYPE_MARKDOWN`

## File Download

### Download and Decrypt Files from Messages

WeWork encrypts files (images, videos, files, etc.) in messages. You need to download and decrypt them using the provided `AesKey`:

```go
apiClient := sdk.NewApiClient(nil)

client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
    // Handle image messages
    if msg.Body.MsgType == sdk.MSG_TYPE_IMAGE {
        if msg.Body.Image.Url != "" && msg.Body.Image.AesKey != "" {
            // Download encrypted file
            resp, err := apiClient.DownloadFileRaw(msg.Body.Image.Url)
            if err != nil {
                return fmt.Errorf("download error: %w", err)
            }

            // Decrypt file data
            imageData, err := sdk.DecryptFile(resp.FileData, msg.Body.Image.AesKey)
            if err != nil {
                return fmt.Errorf("decrypt error: %w", err)
            }

            // Save to file
            f, err := os.Create(resp.FileName)
            if err != nil {
                return fmt.Errorf("create file error: %w", err)
            }
            defer f.Close()

            _, err = f.Write(imageData)
            if err != nil {
                return fmt.Errorf("write file error: %w", err)
            }

            fmt.Printf("Saved image: %s\n", resp.FileName)
        }
    }

    // Similarly for files, videos, etc.
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

### Download Plain Files

For non-encrypted files or external URLs:

```go
apiClient := sdk.NewApiClient(nil)

response, err := apiClient.DownloadFileRaw("https://example.com/file")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Filename: %s\n", response.FileName)
fmt.Printf("Size: %d bytes\n", len(response.FileData))

// Save to file
err = os.WriteFile(response.FileName, response.FileData, 0644)
```

## Error Handling

The SDK distinguishes between recoverable and unrecoverable errors:

- **Recoverable Errors**: Network issues, temporary disconnections → triggers auto-reconnect
- **Unrecoverable Errors**: Authentication failures, JSON marshal errors → stops reconnection

```go
var (
    ErrInvalidConfig      = errors.New("invalid config")
    ErrMissingCredentials = errors.New("missing bot_id or secret")
    ErrNotConnected       = errors.New("not connected")
    ErrAlreadyStarted     = errors.New("client already started")
    ErrAuthFailed         = errors.New("authentication failed")
    ErrAuthTimeout        = errors.New("authentication timeout")
    ErrUnrecoverable      = errors.New("unrecoverable error")
)
```

## Testing

```bash
# Run all tests
go test -v

# Run with race detector
go test -race -v

# Run concurrent tests only
go test -race -v -run TestConcurrent

# Run with coverage
go test -cover
```

The SDK includes comprehensive tests:
- Unit tests for all components
- Concurrency tests (race condition detection)
- Deadlock detection tests
- Connection lifecycle tests

## Architecture

The client uses a multi-goroutine architecture with dual-context lifecycle management:

- **readLoop**: Receives messages from WebSocket
- **writeLoop**: Sends queued messages to WebSocket
- **heartbeatLoop**: Sends periodic ping messages
- **reconnectionLoop**: Handles automatic reconnection

Thread safety is achieved through:
- `connMu`: Protects connection and state
- `callbackMu`: Protects callback functions
- `ctx` + `workerCtx`: Dual-context for graceful shutdown

## Examples

See the [examples](examples/) directory for complete working examples:

- [client_example.go](examples/client_example.go) - Callback, channel, and hybrid modes

## Best Practices

### Debugging and Logging

Use the raw message handler to view all message data, including internal messages like auth, ping, and disconnect events:

```go
// For debugging: log all raw messages
client.SetRawMessageHandler(func(data []byte) error {
    log.Printf("[Raw Message] %s", string(data))
    return nil
})

// More complete example: differentiate message types
client.SetRawMessageHandler(func(data []byte) error {
    var msg map[string]interface{}
    if err := json.Unmarshal(data, &msg); err != nil {
        log.Printf("[Raw Message - Parse Error] %s", string(data))
        return nil
    }

    // Log based on message type
    if cmd, ok := msg["cmd"].(string); ok {
        switch cmd {
        case "aibot_subscribe":
            log.Printf("[Auth Response] %s", string(data))
        case "ping":
            log.Printf("[Ping Response] %s", string(data))
        case "aibot_event_callback":
            // Check if it's a disconnect event
            if body, ok := msg["body"].(map[string]interface{}); ok {
                if event, ok := body["event"].(map[string]interface{}); ok {
                    if eventType, ok := event["eventtype"].(string); ok && eventType == "disconnected_event" {
                        log.Printf("[Disconnect Event] %s", string(data))
                    }
                }
            }
        default:
            log.Printf("[Business Message] cmd=%s", cmd)
        }
    }

    return nil
})
```

**Notes**:

- `SetRawMessageHandler` receives **ALL** messages, including internal ones
- Use only for debugging; in production, disable or log only critical information
- Errors in the raw message handler are not reported via `SetErrorHandler` - handle them yourself

## Requirements

- Go 1.23.12 or later
- Dependencies:
  - `github.com/gorilla/websocket v1.5.3`
  - `github.com/google/uuid v1.6.0`

## License

MIT

## Contributing

Contributions are welcome! Please ensure:
- All tests pass (`go test -race -v`)
- Code is properly formatted (`go fmt`)
- No race conditions detected
- New features include tests

## Reference

Based on the official WeWork AI Bot Node.js SDK: https://github.com/WecomTeam/aibot-node-sdk and doc: https://developer.work.weixin.qq.com/document/path/101463
