package main

import (
	"context"
	"fmt"
	"github.com/JonathanKoerber/CityUCapstoneMSCS/honeypot-core/app/emulator"
	"github.com/JonathanKoerber/CityUCapstoneMSCS/honeypot-core/app/server"
	"github.com/ollama/ollama/api"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
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
	// -------------------------- Emulator Context ----------------------------
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
	sshContext, err := sshEmulator.GetContext()
	if err != nil {
		log.Fatalf("Failed to get SSH context: %v", err)
	}
	err = emulator.EmbedContext(sshContext, store)

	if err != nil {
		log.Fatalf("Failed to embed context: %v", err)
	}
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
		log.Println("Done from incoming Channels")
	}()
	if err != nil {
		log.Fatalf("Failed to create server ssh server: %v", err)
	}
	//ics := ics_node.NewICS()
	//modbusDevice := ics_node.DeviceConfig{
	//	ImageName:      "blink",
	//	IP:             "172.18.0.6",
	//	Port:           "502", // "502"
	//	Net:            "tcp", // "tcp"
	//	BridgeName:     "honeynet",
	//	Protocol:       "modbus",
	//	DeviceName:     "blink_light",
	//	ContextDir:     "",
	//	DockerfilePath: "plc-node/Dockerfile-Modbus-TCP",
	//	Dockerfile:     "Dockerfile-Modbus-TCP",
	//	ContainerName:  "blink_light",
	//}
	// err = ics.BuildAndRunContainer(ctx, modbusDevice)
	// if err != nil { log.Fatalf("Failed to build container: %v", err) }
	fmt.Println("Servers running...")
	select {}
}
