package main

import (
	"fmt"
	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/server"
)

func main() {

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
	select {}
}
