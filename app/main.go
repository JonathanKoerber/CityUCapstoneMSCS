package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/emulator"
	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/server"
	"github.com/ollama/ollama/api"
)

func main() {
	var (
		False = false
		TRUE  = true
	)
	ctx := context.Background()

	var ollamaRawUrl string
	if ollamaRawUrl = os.Getenv("OLLAMA_HOST"); ollamaRawUrl == "" {
		ollamaRawUrl = "http://localhost:11434"
	}

	url, _ := url.Parse(ollamaRawUrl)

	client := api.NewClient(url, http.DefaultClient)

	req := &api.GenerateRequest{
		Model:  "qwen2.5:0.5b",
		Prompt: "The best pizza in the world is",
		Options: map[string]interface{}{
			"temperature":   0.8,
			"repeat_last_n": 2,
		},
		Stream: &TRUE,
	}
	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fmt.Print(resp.Response)
		return nil
	})

	if err != nil {
		log.Fatalln("😡", err)
	}
	fmt.Println()
	// Todo code that kinda works.
	protoclTCPServer := make(map[int]server.ProtocolTCPServer)
	protoclTCPServer[2222] = &server.SSHServer{}
	protoclTCPServer[502] = &server.ModbusTCPServer{}
	// start servers
	// Todo: create logic to start only some services
	for port, server := range protoclTCPServer {
		go server.Start(port)
	}
	defer func() {
		for _, server := range protoclTCPServer {
			server.Stop()
		}
	}()
	fmt.Println("Servers running...")

	ssh := emulator.NewSSHEmulator()
	sshConfig := emulator.NodeConfig{
		ServiceName:  "honeypot",
		Role:         "Client",
		User:         "admin",
		Port:         2221,
		Protocol:     "tcp",
		Service:      "SSH",
		SystemPrompt: "You are emulating a Windows SCADA system. Respond accordingly.",
	}
	ssh.Init(sshConfig)
	// Init Emulator
	if err := ssh.Init(sshConfig); err != nil {
		log.Fatalf("Failed to init SSH emulator: %v", err)
	}

	store := emulator.QdrantStore{
		VectorDBURI: "qdrant",
		Port:        6334,
		OllamaURI:   "http://ollama:11434",
		Model:       "deepseek-coder",
	}
	if err := store.Init([]emulator.Emulator{ssh}); err != nil {
		log.Fatalf("Failed to init store: %v", err)
	}
	points, err := store.VectorContext(ssh)
	if err != nil {
		log.Fatalf("Failed to create vectors: %v", err)
	}
	ctx, err := ssh.GetContext()
	if err != nil {
		log.Fatalf("Failed to get context: %v", err)
	}
	store.AddVectors(ctx.CollectionName, points)
}
