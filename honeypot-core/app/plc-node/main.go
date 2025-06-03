package main

import (
	"log"

	"main/modbusServer"
)

func main() {
	log.Printf("Starting Modbus TCP Server")
	ms := modbusServer.NewModbusTCPServer()
	ms.Start(502)

	log.Printf("Server Modbus running ...")
	// get the server running.
	select {}
}
