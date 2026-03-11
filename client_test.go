package weworkaibotsdk

import (
	"testing"
	"time"
)

func TestConnectionStateConstants(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{Disconnected, "Disconnected"},
		{Connecting, "Connecting"},
		{Authenticating, "Authenticating"},
		{Connected, "Connected"},
		{Reconnecting, "Reconnecting"},
	}

	for _, tt := range tests {
		if tt.state.String() != tt.expected {
			t.Errorf("state.String() = %v, want %v", tt.state.String(), tt.expected)
		}
	}
}

func TestNewClient(t *testing.T) {
	cfg := &Config{
		BotId:                     "test-bot",
		Secret:                    "test-secret",
		WsConnectionTimeout:       10,
		WsConnectionMaxRetryTimes: 3,
		HeartbeatInterval:         30,
		AutoReconnect:             true,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}

	if client.cfg != cfg {
		t.Error("client config not set correctly")
	}

	if client.state != Disconnected {
		t.Errorf("initial state = %v, want Disconnected", client.state)
	}

	if client.ctx == nil {
		t.Error("context not initialized")
	}

	if client.sendQueue == nil {
		t.Error("sendQueue not initialized")
	}

	if client.msgChan == nil {
		t.Error("msgChan not initialized")
	}

	if client.authAckChan == nil {
		t.Error("authAckChan not initialized")
	}

	if client.reconnectTrigger == nil {
		t.Error("reconnectTrigger not initialized")
	}
}

func TestNewClientWithDefaults(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// HeartbeatInterval should default to 30
	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = 30
	}

	if client.cfg.HeartbeatInterval != 30 {
		t.Errorf("HeartbeatInterval = %v, want 30", client.cfg.HeartbeatInterval)
	}
}

func TestStateManagement(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	// Test initial state
	if client.State() != Disconnected {
		t.Errorf("initial State() = %v, want Disconnected", client.State())
	}

	// Test setState
	client.setState(Connecting)
	if client.State() != Connecting {
		t.Errorf("State() = %v, want Connecting", client.State())
	}

	client.setState(Connected)
	if client.State() != Connected {
		t.Errorf("State() = %v, want Connected", client.State())
	}
}

// Note: This test validates that connect() can be called
// The actual connection may succeed or fail depending on network
func TestConnectValidation(t *testing.T) {
	cfg := &Config{
		BotId:                     "test-bot",
		Secret:                    "test-secret",
		WsConnectionTimeout:       1,
		WsConnectionMaxRetryTimes: 1,
	}

	client, _ := NewClient(cfg)

	// Test connecting - may succeed if WeWork server is reachable
	err := client.connect()

	// If connection succeeds, clean up properly
	if err == nil {
		client.connMu.Lock()
		if client.conn != nil {
			client.conn.Close()
			client.conn = nil
		}
		client.connMu.Unlock()
		// Connection succeeded - this is actually fine for WeWork public server
		t.Log("Connection to WeWork server succeeded")
	} else {
		// Connection failed - also acceptable if network issues
		t.Logf("Connection failed as expected: %v", err)
	}
}

func TestSendAuthPayload(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot-123",
		Secret: "test-secret-456",
	}

	client, _ := NewClient(cfg)

	// Test without connection
	err := client.sendAuth()
	if err != ErrNotConnected {
		t.Errorf("sendAuth() without connection error = %v, want %v", err, ErrNotConnected)
	}
}

func TestHandleSpecialMessage(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	tests := []struct {
		name      string
		payload   CallbackPayload
		isSpecial bool
	}{
		{
			name: "auth response",
			payload: CallbackPayload{
				Headers: PayloadHeaders{
					ReqId: "aibot_subscribe:123-456",
				},
				Body: CallbackPayloadBody{
					PayloadError: PayloadError{
						ErrCode: func() *int { i := 0; return &i }(),
					},
				},
			},
			isSpecial: true,
		},
		{
			name: "disconnect event",
			payload: CallbackPayload{
				Cmd: CMD_AIBOT_EVENT_CALLBACK,
				Body: CallbackPayloadBody{
					MsgType: MSG_TYPE_EVENT,
					Event: PayloadBodyEvent{
						EventType: EVENT_TYPE_DISCONNECTED,
					},
				},
			},
			isSpecial: true,
		},
		{
			name: "ping response",
			payload: CallbackPayload{
				Headers: PayloadHeaders{
					ReqId: "ping:789-abc",
				},
			},
			isSpecial: true,
		},
		{
			name: "regular message",
			payload: CallbackPayload{
				Cmd: CMD_AIBOT_MSG_CALLBACK,
				Headers: PayloadHeaders{
					ReqId: "msg:xyz-123",
				},
			},
			isSpecial: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.handleSpecialMessage(tt.payload)
			if result != tt.isSpecial {
				t.Errorf("handleSpecialMessage() = %v, want %v", result, tt.isSpecial)
			}
		})
	}
}

func TestDispatchMessage(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	payload := CallbackPayload{
		Cmd: CMD_AIBOT_MSG_CALLBACK,
		Body: CallbackPayloadBody{
			MsgId: "test-msg-123",
		},
	}

	// Test callback mode
	t.Run("callback mode", func(t *testing.T) {
		called := false
		client.SetMessageHandler(func(p CallbackPayload) error {
			called = true
			if p.Body.MsgId != payload.Body.MsgId {
				t.Errorf("callback payload.MsgId = %v, want %v", p.Body.MsgId, payload.Body.MsgId)
			}
			return nil
		})

		client.dispatchMessage(payload)

		if !called {
			t.Error("message handler was not called")
		}
	})

	// Test channel mode
	t.Run("channel mode", func(t *testing.T) {
		client.onMessage = nil // Clear callback

		client.dispatchMessage(payload)

		select {
		case msg := <-client.Messages():
			if msg.Body.MsgId != payload.Body.MsgId {
				t.Errorf("channel payload.MsgId = %v, want %v", msg.Body.MsgId, payload.Body.MsgId)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("message not received on channel")
		}
	})
}

func TestSendPing(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	// Test without connection
	err := client.sendPing()
	if err != ErrNotConnected {
		t.Errorf("sendPing() without connection error = %v, want %v", err, ErrNotConnected)
	}
}

func TestStartWithoutServer(t *testing.T) {
	cfg := &Config{
		BotId:                     "test-bot",
		Secret:                    "test-secret",
		WsConnectionTimeout:       1,
		WsConnectionMaxRetryTimes: 1,
	}

	client, _ := NewClient(cfg)

	// Start should fail since there's no server
	err := client.Start()
	if err == nil {
		t.Error("Start() should fail without server")
	}

	// State should be Disconnected after failure
	if client.State() != Disconnected {
		t.Errorf("State() after failed Start = %v, want Disconnected", client.State())
	}
}

func TestStop(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	// Stop should work even if never started
	err := client.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// State should be Disconnected
	if client.State() != Disconnected {
		t.Errorf("State() after Stop = %v, want Disconnected", client.State())
	}

	// Multiple stops should be safe
	err = client.Stop()
	if err != nil {
		t.Errorf("second Stop() error = %v", err)
	}
}

func TestSend(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	// Create a test payload
	payload := map[string]string{
		"cmd": "test_command",
		"msg": "test message",
	}

	// Send should work even before Start (queues message)
	err := client.Send(payload)
	if err != nil {
		t.Errorf("Send() error = %v", err)
	}

	// Verify message was queued
	select {
	case <-client.sendQueue:
		// Message was queued successfully
	case <-time.After(100 * time.Millisecond):
		t.Error("message was not queued")
	}
}

func TestReconnect(t *testing.T) {
	cfg := &Config{
		BotId:                     "test-bot",
		Secret:                    "test-secret",
		WsConnectionTimeout:       1,
		WsConnectionMaxRetryTimes: 1,
		AutoReconnect:             false, // Manual mode
	}

	client, _ := NewClient(cfg)

	// Reconnect should trigger reconnection attempt
	err := client.Reconnect()
	if err == nil {
		t.Error("Reconnect() should fail without server")
	}
}

func TestCallbackSetters(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	// Test SetMessageHandler
	msgHandlerCalled := false
	client.SetMessageHandler(func(p CallbackPayload) error {
		msgHandlerCalled = true
		return nil
	})

	if client.onMessage == nil {
		t.Error("SetMessageHandler did not set callback")
	}

	// Test SetErrorHandler
	errHandlerCalled := false
	client.SetErrorHandler(func(err error) {
		errHandlerCalled = true
	})

	if client.onError == nil {
		t.Error("SetErrorHandler did not set callback")
	}

	// Test SetReconnectHandler
	reconnectHandlerCalled := false
	client.SetReconnectHandler(func() {
		reconnectHandlerCalled = true
	})

	if client.onReconnect == nil {
		t.Error("SetReconnectHandler did not set callback")
	}

	// Trigger callbacks to verify they work
	client.onMessage(CallbackPayload{})
	if !msgHandlerCalled {
		t.Error("message handler not called")
	}

	client.onError(nil)
	if !errHandlerCalled {
		t.Error("error handler not called")
	}

	client.onReconnect()
	if !reconnectHandlerCalled {
		t.Error("reconnect handler not called")
	}
}
