package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	sdk "wework_aibot_sdk"

	"github.com/google/uuid"
)

func main() {
	// Example 1: Callback mode
	// callbackMode()

	// Example 2: Channel mode
	// channelMode()

	// Example 3: Hybrid mode
	hybridMode()
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
	apiClient := sdk.NewApiClient(nil)
	chatId := ""
	// Handle raw messages in callback
	client.SetRawMessageHandler(func(msg []byte) error {
		fmt.Printf("+++++ raw msg: %s\n\n", msg)
		return nil
	})
	client.SetMessageHandler(func(msg sdk.CallbackPayload) error {
		if msg.Body.MsgType == sdk.MSG_TYPE_TEXT {
			fmt.Printf("----- text message: %s\n\n", msg.Body.Text.Content)
			return nil
		}
		fmt.Printf("======%#v\n\n", msg)
		return nil
	})

	// Process all messages asynchronously
	go func() {
		for msg := range client.Messages() {
			fmt.Printf(">>>>>Async processing: %+v\n\n", msg)
			if msg.Body.ChatId != "" {
				chatId = msg.Body.ChatId
			}
			if msg.Body.From.UserId != "" {
				chatId = msg.Body.From.UserId
			}
			if msg.Body.MsgType == sdk.MSG_TYPE_IMAGE {
				if msg.Body.Image.Url != "" && msg.Body.Image.AesKey != "" {
					resp, err := apiClient.DownloadFileRaw(msg.Body.Image.Url)
					if err != nil {
						fmt.Printf("DownloadFileRaw error: %s\n", err)
						continue
					}
					imageData, err := sdk.DecryptFile(resp.FileData, msg.Body.Image.AesKey)
					if err != nil {
						fmt.Printf("DecryptFile error: %s\n", err)
						continue
					}
					f, err := os.Create(resp.FileName)
					if err != nil {
						fmt.Printf("Create file error: %s\n", err)
						continue
					}
					defer f.Close()
					_, err = f.Write(imageData)
					if err != nil {
						fmt.Printf("Write file error: %s\n", err)
						continue
					}
				}
			}
		}
	}()

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	fmt.Println("Client started successfully!")
	fmt.Println("Type your message and press Enter to send.")
	fmt.Println("Type '.exit' to quit.")

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Create scanner for user input
	scanner := bufio.NewScanner(os.Stdin)

	// Input loop
	go func() {
		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				return
			}

			input := strings.TrimSpace(scanner.Text())

			// Check for exit command
			if input == ".exit" {
				fmt.Println("Exiting...")
				sigChan <- os.Interrupt
				return
			}

			// Skip empty input
			if input == "" {
				continue
			}
			if chatId == "" {
				fmt.Println("please send a message to bot first to get chatId or userId")
				continue
			}

			// Send markdown message via client
			payload := sdk.PushPayload{
				Cmd: sdk.CMD_AIBOT_SEND_MSG,
				Headers: sdk.PayloadHeaders{
					ReqId: uuid.New().String(),
				},
				Body: sdk.PushPayloadBody{
					MsgType: sdk.MSG_TYPE_MARKDOWN,
					ChatId:  chatId,
					Markdown: &sdk.PayloadBodyMarkdown{
						Content: input,
					},
				},
			}

			if err := client.Send(payload); err != nil {
				fmt.Printf("Failed to send message: %v\n", err)
			} else {
				fmt.Println("Message sent!")
			}
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down...")
}
