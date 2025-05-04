package main

import (
	"fmt"
	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/server"
)

func main() {

	protoclTCPServer := make(map[int]server.ProtocolTCPServer)
	protoclTCPServer[2222] = &server.SSHServer{}

	// start servers
	// Todo: create logic to start only some services
	for port, server := range protoclTCPServer {
		server.Start(port)
	}
	defer func() {
		for _, server := range protoclTCPServer {
			server.Stop()
		}
	}()
	fmt.Println("Servers running...")
	select {}
}
