package weworkaibotsdk

import (
	"sync"
	"testing"
	"time"
)

// TestConcurrentStateAccess tests concurrent state access doesn't deadlock
func TestConcurrentStateAccess(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	// Test concurrent state reads and writes
	var wg sync.WaitGroup
	iterations := 100

	// Multiple goroutines reading state
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = client.State()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Multiple goroutines writing state
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			states := []ConnectionState{Disconnected, Connecting, Connected, Reconnecting}
			for j := 0; j < iterations; j++ {
				client.setState(states[j%len(states)])
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Wait with timeout to detect deadlock
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(5 * time.Second):
		t.Fatal("Deadlock detected: concurrent state access timed out")
	}
}

// TestConcurrentSendQueue tests concurrent Send operations
func TestConcurrentSendQueue(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	var wg sync.WaitGroup
	iterations := 50

	// Multiple goroutines sending messages
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				payload := map[string]interface{}{
					"id":   id,
					"iter": j,
				}
				client.Send(payload)
			}
		}(i)
	}

	// Drain the queue in another goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations*10; i++ {
			select {
			case <-client.sendQueue:
			case <-time.After(10 * time.Millisecond):
				return
			}
		}
	}()

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Deadlock detected: concurrent send operations timed out")
	}
}

// TestStopWhileSending tests that Stop doesn't deadlock when messages are being sent
func TestStopWhileSending(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	// Start sending messages
	go func() {
		for i := 0; i < 1000; i++ {
			client.Send(map[string]int{"count": i})
			time.Sleep(time.Microsecond * 100)
		}
	}()

	// Stop after a short delay
	time.Sleep(10 * time.Millisecond)

	done := make(chan error, 1)
	go func() {
		done <- client.Stop()
	}()

	select {
	case <-done:
		// Success - Stop completed
	case <-time.After(2 * time.Second):
		t.Fatal("Deadlock detected: Stop() timed out while messages being sent")
	}
}

// TestRapidStartStop tests rapid Start/Stop cycles
func TestRapidStartStop(t *testing.T) {
	cfg := &Config{
		BotId:                     "test-bot",
		Secret:                    "test-secret",
		WsConnectionTimeout:       1,
		WsConnectionMaxRetryTimes: 1,
	}

	for i := 0; i < 5; i++ {
		client, _ := NewClient(cfg)

		// Try to start (will fail without server)
		go client.Start()

		// Immediately try to stop
		time.Sleep(5 * time.Millisecond)

		done := make(chan error, 1)
		go func() {
			done <- client.Stop()
		}()

		select {
		case <-done:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatalf("Deadlock detected on iteration %d: rapid Start/Stop", i)
		}
	}
}

// TestConcurrentCallbackSet tests concurrent callback setter access
func TestConcurrentCallbackSet(t *testing.T) {
	cfg := &Config{
		BotId:  "test-bot",
		Secret: "test-secret",
	}

	client, _ := NewClient(cfg)

	var wg sync.WaitGroup

	// Concurrent SetMessageHandler
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				client.SetMessageHandler(func(p CallbackPayload) error {
					return nil
				})
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Concurrent SetErrorHandler
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				client.SetErrorHandler(func(err error) {})
				time.Sleep(time.Microsecond)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Deadlock detected: concurrent callback setting timed out")
	}
}
