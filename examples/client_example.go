package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	sdk "wework_aibot_sdk"
)

func main() {
	// Example 1: Callback mode
	callbackMode()

	// Example 2: Channel mode
	// channelMode()

	// Example 3: Hybrid mode
	// hybridMode()
}

func callbackMode() {
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
		fmt.Printf("Received message: %+v\n", msg)
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

	fmt.Println("Shutting down...")
}

func channelMode() {
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

func hybridMode() {
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

	// Handle critical messages in callback
	client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
		if msg.Body.MsgType == sdk.MSG_TYPE_TEXT {
			fmt.Printf("Critical text message: %s\n", msg.Body.Text.Content)
		}
		return nil
	})

	// Process all messages asynchronously
	go func() {
		for msg := range client.Messages() {
			fmt.Printf("Async processing: %+v\n", msg)
		}
	}()

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
