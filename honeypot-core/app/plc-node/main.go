package main

import (
	"log"
	"time"

	"main/modbusServer"
)

func main() {
	log.Printf("Starting Modbus TCP Server")
	server, handler := modbusServer.NewModbusTCPServer(502)
	server.Start()
	log.Printf("Server Modbus running ...")
	// get the server running.

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for _, device := range handler.Device {

		for {
			select {
			case <-ticker.C:
				device.SimulateActivity()
			}
		}
		select {}
	}
}
