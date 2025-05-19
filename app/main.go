package main

import (
	"context"
	"fmt"
	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/server"
	"github.com/qdrant/go-client/qdrant"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/emulator"
	"github.com/ollama/ollama/api"
)

// TODO move env to docker env
func main() {
	// ENV VAR
	os.Setenv("GO_ENV", "development")
	os.Setenv("OLLAMA_URL", "http://Ollama:11434")
	os.Setenv("MODEL", "mistral")
	//os.Setenv("VEC_STORE_URI", "mongodb://admin:password@vector_store:27017/?authSource=admin")
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

	store := emulator.Store{}
	if err := store.Init(); err != nil {
		log.Printf("Failed to init store: %v", err)
	}
	// set up store
	// register emulators
	sshEmulator := emulator.NewSSHEmulator()
	if err := sshEmulator.Init(&store); err != nil {
		log.Fatalf("Failed to init ssh emulator: %v", err)
	}
	sshContext, _ := sshEmulator.GetContext()
	lineChan := make(chan string, 100)
	embeddingChan := make(chan *qdrant.PointStruct, 1000)
	log.Printf("Starting to embed data")
	go func() {
		defer close(lineChan)
		lines, err := store.ReadContextFiles(sshContext)
		if err != nil {
			log.Printf("Failed to read context files: %v", err)
		}
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
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
		defer close(embeddingChan)
		for line := range lineChan {
			embedding, err := store.EmbedDocs(line)
			if err != nil {
				log.Printf("Failed to embed context: %v", err)
				continue
			}
			select {
			case embeddingChan <- embedding:
			default:
				log.Printf("Embedding channel full, waiting...")
				time.Sleep(1 * time.Second)
			}
		}
	}()
	go func() {
		batch := make([]*qdrant.PointStruct, 0, 100)
		ticker := time.NewTicker(2 * time.Second)
		batchSize := 20
		defer ticker.Stop()
		for {
			select {
			case embedding, ok := <-embeddingChan:
				if !ok {
					if len(batch) > 0 {
						store.AddVectors("ssh_emulator", batch)
					}
					log.Printf("Ebedding err: %v", ok)
					return
				}
				if embedding != nil {
					batch = append(batch, embedding)
				}
				batch = append(batch, embedding)
				if len(batch) > batchSize {
					store.AddVectors("ssh_emulator", batch)
					batch = batch[:0]
				}
			case <-ticker.C:
				if len(batch) > 0 {
					store.AddVectors("ssh_emulator", batch)
					batch = batch[:0]
				}
			}
		}
	}()
	// Start and run protocol Servers
	log.Println("Data embedded")
	sshServer, err := server.NewSSHServer(2222)
	if err != nil {
		log.Fatalf("Failed init server: %v", err)
	}
	sshServer.Start()
	log.Println("Server started getting ready to accept connections")
	inComingChan := make(chan net.Conn, 100)
	go func() {
		log.Println("Listening for incoming connections")
		for {
			log.Println("Waiting for incoming connection")
			inComing, err := sshServer.Listener.Accept()
			if err != nil {
				log.Printf("Failed to accept incomming socket: %v", err)
				continue
			}
			inComingChan <- inComing
		}
		log.Println("Closing incoming connections")
	}()
	go func() {
		log.Println("Listening for outgoing connections from inComming Channels")
		sshEmulator := emulator.NewSSHEmulator()
		sshEmulator.Init(&store)
		for conn := range inComingChan {
			go sshServer.HandleConn(conn, *sshEmulator)
		}
		log.Println("Done from inComming Channels")
	}()
	if err != nil {
		log.Fatalf("Failed to create server ssh server: %v", err)
	}

	fmt.Println("Servers running...")
	select {}
}
