package weworkaibotsdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ConnectionState represents the current state of the WebSocket connection
type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Authenticating
	Connected
	Reconnecting
)

// String returns the string representation of ConnectionState
func (s ConnectionState) String() string {
	switch s {
	case Disconnected:
		return "Disconnected"
	case Connecting:
		return "Connecting"
	case Authenticating:
		return "Authenticating"
	case Connected:
		return "Connected"
	case Reconnecting:
		return "Reconnecting"
	default:
		return "Unknown"
	}
}

var (
	ErrInvalidConfig      = errors.New("invalid config")
	ErrMissingCredentials = errors.New("missing bot_id or secret")
	ErrNotConnected       = errors.New("not connected")
	ErrAlreadyStarted     = errors.New("client already started")
	ErrAuthFailed         = errors.New("authentication failed")
	ErrAuthTimeout        = errors.New("authentication timeout")
	ErrUnrecoverable      = errors.New("unrecoverable error")
)

// Client is a WebSocket client with auto-reconnection and dual message delivery
type Client struct {
	cfg *Config

	// Connection state
	conn   *websocket.Conn
	connMu sync.RWMutex
	state  ConnectionState

	// Goroutine coordination
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Worker goroutines lifecycle management
	workerCtx    context.Context
	workerCancel context.CancelFunc
	workerWg     sync.WaitGroup

	// Channels
	sendQueue        chan []byte
	msgChan          chan CallbackPayload
	authAckChan      chan error
	reconnectTrigger chan struct{}

	// Callbacks (protected by callbackMu)
	callbackMu  sync.RWMutex
	onMessage   func(CallbackPayload) error
	onError     func(error)
	onReconnect func()
}

// NewClient creates a new WebSocket client
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}

	if cfg.BotId == "" || cfg.Secret == "" {
		return nil, ErrMissingCredentials
	}

	// Set defaults
	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = 30
	}

	if cfg.WsConnectionTimeout == 0 {
		cfg.WsConnectionTimeout = 10
	}

	if cfg.WsConnectionMaxRetryTimes == 0 {
		cfg.WsConnectionMaxRetryTimes = 3
	}

	if cfg.WsConnectionRetryInterval == 0 {
		cfg.WsConnectionRetryInterval = 2
	}

	if cfg.SendQueueSize == 0 {
		cfg.SendQueueSize = 100
	}

	if cfg.MsgChanSize == 0 {
		cfg.MsgChanSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())
	workerCtx, workerCancel := context.WithCancel(ctx)

	return &Client{
		cfg:              cfg,
		state:            Disconnected,
		ctx:              ctx,
		cancel:           cancel,
		workerCtx:        workerCtx,
		workerCancel:     workerCancel,
		sendQueue:        make(chan []byte, cfg.SendQueueSize),
		msgChan:          make(chan CallbackPayload, cfg.MsgChanSize),
		authAckChan:      make(chan error, 1),
		reconnectTrigger: make(chan struct{}, 1),
	}, nil
}

// State returns the current connection state (thread-safe)
func (c *Client) State() ConnectionState {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.state
}

// setState updates the connection state (internal use only)
func (c *Client) setState(state ConnectionState) {
	c.connMu.Lock()
	c.state = state
	c.connMu.Unlock()
}

// connect establishes WebSocket connection with retry logic
func (c *Client) connect() error {
	retryCount := 0

	for retryCount < c.cfg.WsConnectionMaxRetryTimes {
		c.setState(Connecting)

		ctx, cancel := context.WithTimeout(context.Background(),
			time.Duration(c.cfg.WsConnectionTimeout)*time.Second)

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, WEBSOCKET_URL, nil)
		cancel()

		if err == nil {
			c.connMu.Lock()
			c.conn = conn
			c.connMu.Unlock()
			return nil
		}

		retryCount++
		if retryCount < c.cfg.WsConnectionMaxRetryTimes {
			time.Sleep(time.Duration(c.cfg.WsConnectionRetryInterval) * time.Second)
		}
	}

	c.setState(Disconnected)
	return fmt.Errorf("failed to connect after %d retries", c.cfg.WsConnectionMaxRetryTimes)
}

// sendAuth sends authentication payload to server
func (c *Client) sendAuth() error {
	c.connMu.RLock()
	conn := c.conn
	c.connMu.RUnlock()

	if conn == nil {
		return ErrNotConnected
	}

	authPayload := AuthPayload{
		Cmd: CMD_AIBOT_SUBSCRIBE,
		Headers: PayloadHeaders{
			ReqId: string(CMD_AIBOT_SUBSCRIBE) + ":" + uuid.NewString(),
		},
		Body: AuthPayloadBody{
			BotId:  c.cfg.BotId,
			Secret: c.cfg.Secret,
		},
	}

	authPayloadData, err := json.Marshal(authPayload)
	if err != nil {
		return fmt.Errorf("%w: marshal auth payload: %w", ErrUnrecoverable, err)
	}

	return conn.WriteMessage(websocket.TextMessage, authPayloadData)
}

// authenticate sends auth and waits for confirmation
func (c *Client) authenticate() error {
	c.setState(Authenticating)

	if err := c.sendAuth(); err != nil {
		c.setState(Disconnected)
		return fmt.Errorf("send auth: %w", err)
	}

	// Wait for auth response with timeout
	timeout := time.Duration(c.cfg.WsHandshakeTimeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	select {
	case err := <-c.authAckChan:
		if err != nil {
			c.setState(Disconnected)
			return err
		}
		c.setState(Connected)
		return nil
	case <-time.After(timeout):
		c.setState(Disconnected)
		return ErrAuthTimeout
	case <-c.ctx.Done():
		c.setState(Disconnected)
		return c.ctx.Err()
	}
}

// handleSpecialMessage processes auth, disconnect, and ping messages
// Returns true if message was handled (should not dispatch to user)
func (c *Client) handleSpecialMessage(payload CallbackPayload) bool {
	reqId := payload.Headers.ReqId
	// Handle auth response
	if strings.HasPrefix(reqId, string(CMD_AIBOT_SUBSCRIBE)) {
		if payload.ErrCode != nil && *payload.ErrCode == ERROR_CODE_SUCCESS {
			c.authAckChan <- nil
		} else {
			errMsg := payload.ErrMsg
			if errMsg == "" {
				errMsg = "authentication failed"
			}
			c.authAckChan <- fmt.Errorf("%w: %s", ErrAuthFailed, errMsg)
		}
		return true
	}

	// Handle disconnect event
	if payload.Cmd == CMD_AIBOT_EVENT_CALLBACK &&
		payload.Body.MsgType == MSG_TYPE_EVENT &&
		payload.Body.Event.EventType == EVENT_TYPE_DISCONNECTED {
		// Server requested disconnect - clean shutdown
		go c.Stop()
		return true
	}

	// Handle ping response (silent acknowledge)
	if strings.HasPrefix(reqId, string(CMD_PING)) {
		return true
	}

	return false
}

// dispatchMessage sends message to callback and/or channel
func (c *Client) dispatchMessage(payload CallbackPayload) {
	// Get callbacks with lock
	c.callbackMu.RLock()
	onMessage := c.onMessage
	onError := c.onError
	c.callbackMu.RUnlock()

	// Callback mode
	if onMessage != nil {
		if err := onMessage(payload); err != nil && onError != nil {
			onError(fmt.Errorf("message handler error: %w", err))
		}
	}

	// Channel mode (non-blocking send)
	if c.msgChan != nil {
		select {
		case c.msgChan <- payload:
		default:
			// Channel full, message dropped
			// Could log warning here
		}
	}
}

// Messages returns the message channel for channel-based consumption
// Returns nil if client is not initialized
func (c *Client) Messages() <-chan CallbackPayload {
	return c.msgChan
}

// SetMessageHandler sets the callback for incoming messages
func (c *Client) SetMessageHandler(handler func(CallbackPayload) error) {
	c.callbackMu.Lock()
	c.onMessage = handler
	c.callbackMu.Unlock()
}

// SetErrorHandler sets the callback for connection/runtime errors
func (c *Client) SetErrorHandler(handler func(error)) {
	c.callbackMu.Lock()
	c.onError = handler
	c.callbackMu.Unlock()
}

// SetReconnectHandler sets the callback for successful reconnections
func (c *Client) SetReconnectHandler(handler func()) {
	c.callbackMu.Lock()
	c.onReconnect = handler
	c.callbackMu.Unlock()
}

// readLoop reads messages from WebSocket connection
func (c *Client) readLoop() {
	defer c.wg.Done()
	defer c.workerWg.Done()

	for {
		select {
		case <-c.workerCtx.Done():
			return
		default:
		}

		c.connMu.RLock()
		conn := c.conn
		c.connMu.RUnlock()

		if conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			c.handleConnectionError(err)
			return
		}

		var payload CallbackPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			// Skip malformed messages
			continue
		}

		// Handle special messages (auth, disconnect, ping)
		if c.handleSpecialMessage(payload) {
			continue
		}

		// Dispatch to user
		c.dispatchMessage(payload)
	}
}

// handleConnectionError triggers reconnection or reports error
func (c *Client) handleConnectionError(err error) {
	if c.cfg.AutoReconnect {
		// Non-blocking send to reconnect trigger
		select {
		case c.reconnectTrigger <- struct{}{}:
		default:
		}
	} else {
		c.callbackMu.RLock()
		onError := c.onError
		c.callbackMu.RUnlock()

		if onError != nil {
			onError(fmt.Errorf("connection error: %w", err))
		}
	}
}

// writeLoop writes messages from sendQueue to WebSocket
func (c *Client) writeLoop() {
	defer c.wg.Done()
	defer c.workerWg.Done()

	for {
		select {
		case <-c.workerCtx.Done():
			// Worker context cancelled - exit immediately
			return
		case <-c.ctx.Done():
			// Client context cancelled - drain queue before exit
			c.drainSendQueue()
			return
		case msg := <-c.sendQueue:
			c.connMu.RLock()
			conn := c.conn
			c.connMu.RUnlock()

			if conn == nil {
				// Connection closed - exit this worker
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.handleConnectionError(err)
				return
			}
		}
	}
}

// drainSendQueue attempts to send remaining messages before shutdown
func (c *Client) drainSendQueue() {
	for {
		select {
		case msg := <-c.sendQueue:
			c.connMu.RLock()
			conn := c.conn
			c.connMu.RUnlock()

			if conn != nil {
				conn.WriteMessage(websocket.TextMessage, msg)
			}
		default:
			return
		}
	}
}

// sendPing sends heartbeat ping to server
func (c *Client) sendPing() error {
	c.connMu.RLock()
	conn := c.conn
	c.connMu.RUnlock()

	if conn == nil {
		return ErrNotConnected
	}

	pingPayload := PingPayload{
		Cmd: CMD_PING,
		Headers: PayloadHeaders{
			ReqId: string(CMD_PING) + ":" + uuid.NewString(),
		},
	}

	pingPayloadData, err := json.Marshal(pingPayload)
	if err != nil {
		return fmt.Errorf("%w: marshal ping payload: %w", ErrUnrecoverable, err)
	}

	return conn.WriteMessage(websocket.TextMessage, pingPayloadData)
}

// heartbeatLoop sends periodic ping messages
func (c *Client) heartbeatLoop() {
	defer c.wg.Done()
	defer c.workerWg.Done()

	interval := time.Duration(c.cfg.HeartbeatInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.workerCtx.Done():
			return
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.sendPing(); err != nil {
				// If ErrNotConnected, this worker should exit
				if errors.Is(err, ErrNotConnected) {
					return
				}
				// Other ping failures are not critical, connection error will be caught by reader
				continue
			}
		}
	}
}

// reconnectionLoop monitors and handles automatic reconnection
func (c *Client) reconnectionLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.reconnectTrigger:
			c.setState(Reconnecting)

			// Stop worker goroutines gracefully
			c.workerCancel()  // Cancel worker context
			c.workerWg.Wait() // Wait for all workers to exit

			// Close old connection
			c.connMu.Lock()
			if c.conn != nil {
				c.conn.Close()
				c.conn = nil
			}
			c.connMu.Unlock()

			// Attempt reconnection
			if err := c.connect(); err != nil {
				c.callbackMu.RLock()
				onError := c.onError
				c.callbackMu.RUnlock()

				if onError != nil {
					onError(fmt.Errorf("reconnect failed: %w", err))
				}

				// Check if this is an unrecoverable error
				if errors.Is(err, ErrUnrecoverable) {
					c.setState(Disconnected)
					return
				}

				// Stay in reconnecting state, will retry on next trigger
				continue
			}

			// Re-authenticate
			if err := c.authenticate(); err != nil {
				c.callbackMu.RLock()
				onError := c.onError
				c.callbackMu.RUnlock()

				if onError != nil {
					onError(fmt.Errorf("re-authentication failed: %w", err))
				}

				// Check if this is an unrecoverable error
				if errors.Is(err, ErrUnrecoverable) {
					// Unrecoverable error - stop reconnection attempts
					c.setState(Disconnected)
					return
				}

				// Close connection and stay in reconnecting state
				c.connMu.Lock()
				if c.conn != nil {
					c.conn.Close()
					c.conn = nil
				}
				c.connMu.Unlock()
				continue
			}

			// Reconnection successful
			c.setState(Connected)

			// Create new worker context for new goroutines
			c.workerCtx, c.workerCancel = context.WithCancel(c.ctx)

			// Restart reader, writer, and heartbeat goroutines
			c.wg.Add(3)
			c.workerWg.Add(3)
			go c.readLoop()
			go c.writeLoop()
			go c.heartbeatLoop()

			// Notify application
			c.callbackMu.RLock()
			onReconnect := c.onReconnect
			c.callbackMu.RUnlock()

			if onReconnect != nil {
				onReconnect()
			}
		}
	}
}

// Start connects to WebSocket server and authenticates
// Blocks until authentication succeeds or fails
func (c *Client) Start() error {
	// Prevent multiple starts
	if c.State() != Disconnected {
		return ErrAlreadyStarted
	}

	// Set handshake timeout if configured
	if c.cfg.WsHandshakeTimeout > 0 {
		websocket.DefaultDialer.HandshakeTimeout = time.Duration(c.cfg.WsHandshakeTimeout) * time.Second
	}

	// Connect to server
	if err := c.connect(); err != nil {
		return err
	}

	// Start reader goroutine to receive auth response
	c.wg.Add(1)
	c.workerWg.Add(1)
	go c.readLoop()

	// Authenticate
	if err := c.authenticate(); err != nil {
		// Authentication failed - clean up
		c.workerCancel() // Cancel worker context

		c.connMu.Lock()
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.connMu.Unlock()

		// Wait for reader to exit (it will exit when conn is closed or workerCtx is done)
		c.workerWg.Wait()
		return err
	}

	// Authentication successful - start other goroutines
	c.wg.Add(2)
	c.workerWg.Add(2)
	go c.writeLoop()
	go c.heartbeatLoop()

	// Start reconnection loop if auto-reconnect enabled
	if c.cfg.AutoReconnect {
		c.wg.Add(1)
		go c.reconnectionLoop()
	}

	return nil
}

// Stop gracefully shuts down the client
func (c *Client) Stop() error {
	// Check if already stopped
	if c.State() == Disconnected && c.ctx.Err() != nil {
		return nil
	}

	// Cancel worker context first to stop worker goroutines
	c.workerCancel()

	// Cancel main context to signal all goroutines to exit
	c.cancel()

	// Close WebSocket connection BEFORE waiting for goroutines
	// This unblocks any pending ReadMessage() calls in readLoop
	c.connMu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connMu.Unlock()

	// Wait for all goroutines to exit
	c.wg.Wait()

	c.setState(Disconnected)
	return nil
}

// Send queues a message to be sent to the server
// Accepts any JSON-marshalable payload
func (c *Client) Send(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%w: marshal payload: %w", ErrUnrecoverable, err)
	}

	select {
	case c.sendQueue <- data:
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("client stopped")
	}
}
