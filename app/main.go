package main

import (
	"context"
	"fmt"
	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/server"
	"github.com/qdrant/go-client/qdrant"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/emulator"
	"github.com/ollama/ollama/api"
)

func main() {
	// ENV VAR
	os.Setenv("GO_ENV", "development")
	os.Setenv("OLLAMA_URL", "http://Ollama:11434")
	os.Setenv("MODEL", "mistral")
	os.Setenv("VEC_STORE_URL", "qdrant")
	os.Setenv("VEC_STORE_PORT", "6334")

	ctx := context.Background()

	var ollamaRawUrl string
	if ollamaRawUrl = os.Getenv("OLLAMA_URL"); ollamaRawUrl == "" {
		ollamaRawUrl = "http://Ollama:11434"
	}

	ollamaURL, _ := url.Parse(ollamaRawUrl)
	client := api.NewClient(ollamaURL, http.DefaultClient)
	modelName := os.Getenv("MODEL")
	models, err := client.List(ctx)

	log.Printf("Ollama list: %v", models)
	found := false
	if err != nil || models == nil {
		log.Printf("Failed to list models: %v:", err)
	} else {
		for _, m := range models.Models {
			log.Printf("Model name: %s\n", m.Name)
			if m.Name == modelName {
				found = true
				log.Printf("Found model name: %s\n", m.Name)
				break
			}
		}
	}

	progressFunc := func(resp api.ProgressResponse) error {
		fmt.Printf("Progress: status=%v, total=%v, completed=%v\n", resp.Status, resp.Total, resp.Completed)
		return nil
	}

	if !found {
		log.Printf("Model not found: %s", modelName)
		err = client.Pull(ctx, &api.PullRequest{Model: modelName}, progressFunc)

		if err != nil {
			log.Printf("Failed to pull model: %v:", err)
		} else {
			log.Printf("Pulling model successfully: %s\n", modelName)
		}
	} else {
		log.Printf("Model found: %s\n", modelName)
	}
	// DEBUG: check to see if the correct model is running
	// var ( False = false TRUE = true )
	//req := &api.GenerateRequest{
	//	Model:  modelName,
	//	Prompt: "The best pizza in the world is",
	//	Options: map[string]interface{}{
	//		"temperature":   0.8,
	//		"repeat_last_n": 2,
	//	},
	//	Stream: &TRUE,
	//}

	//err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
	//	log.Print(resp.Response)
	//	return nil
	//})
	//
	//if err != nil {
	//	log.Printf("😡 : %v", err)
	//} else {
	//	log.Printf("Ollama generated successfully")
	//	log.Printf("Ollama, response: %v:", req)
	//}

	store := emulator.QdrantStore{
		VectorDBURI: os.Getenv("VEC_STORE_URL"),
		Port:        6334,
		OllamaURI:   os.Getenv("OLLAMA_URL"),
		Model:       os.Getenv("MODEL"),
	}
	// register emulators
	sshEmulator := emulator.NewSSHEmulator()
	if err := sshEmulator.Init(&store); err != nil {
		log.Fatalf("Failed to init ssh emulator: %v", err)
	}
	// set up store
	if err := store.Init([]emulator.Emulator{sshEmulator}); err != nil {
		log.Printf("Failed to init store: %v", err)
	}
	lineChan := make(chan string, 1000)
	embedingsChan := make(chan *qdrant.PointStruct, 1000)

	go func() {
		defer close(lineChan)
		lines, err := store.ReadContextFiles(sshEmulator)
		if err != nil {
			log.Printf("Failed to read context files: %v", err)
		}
		for _, line := range lines {
			select {
			case lineChan <- line:
			default:
				log.Printf("Line channel full, waiting...")
				time.Sleep(1 * time.Second)
				lineChan <- line
			}
		}
	}()
	go func() {
		defer close(embedingsChan)
		for line := range lineChan {
			embedding, err := store.EmbedContext(line)
			if err != nil {
				log.Printf("Failed to embed context: %v", err)
				continue
			}
			select {
			case embedingsChan <- embedding:
			default:
				log.Printf("Embedding channel full, waiting...")
				time.Sleep(1 * time.Second)
			}
		}
	}()
	go func() {
		for embedding := range embedingsChan {
			nodeCtx, err := sshEmulator.GetContext()
			if err != nil {
				log.Printf("Failed to get context: %v", err)
			}
			err = store.AddVectors(nodeCtx.CollectionName, embedding)
			if err != nil {
				log.Printf("Failed to add vectors: %v", err)
			}
		}
	}()
	// Register Protocol Servers
	protocolTCPServer := make(map[int]server.ProtocolTCPServer)
	sshServer, err := server.NewSSHServer()
	if err != nil {
		log.Fatalf("Failed to create server ssh server: %v", err)
	}
	protocolTCPServer[2222] = sshServer

	//protocolTCPServer[502] = &server.ModbusTCPServer{}
	// start servers
	// Todo: create logic to start only some services
	//for port, server := range protocolTCPServer {
	//	log.Printf("Starting server on port %d\n", port)
	//	go server.Start(port)
	//}
	//defer func() {
	//	for _, server := range protocolTCPServer {
	//		server.Stop()
	//	}
	//}()
	sshServer.Start(888)
	fmt.Println("Servers running...")
}
