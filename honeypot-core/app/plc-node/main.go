package main

import (
	"log"

	"main/modbusServer"
)

func main() {
	log.Printf("Starting Modbus TCP Server")
	server := modbusServer.NewModbusTCPServer(502)
	server.Start()
	log.Printf("Server Modbus running ...")
	// get the server running.
	select {}
}
