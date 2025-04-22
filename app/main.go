package main

import (
	"fmt"
	"github.com/JonathanKoerber/CityUCapstoneMSCS/app/server"
	"time"
)

func main() {
	sshServer := server.NewSSHServer(":2222")
	if err := sshServer.Start(); err != nil {
		panic(err)
	}
	for {
		time.Sleep(10 * time.Second)
	}
	fmt.Println("Hello World")
}
