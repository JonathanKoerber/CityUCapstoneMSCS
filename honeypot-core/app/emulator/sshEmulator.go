package emulator

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"log"
)

type SSHEmulator struct {
	Context NodeContext
}

func NewSSHEmulator() *SSHEmulator {
	return &SSHEmulator{}
}
func (s *SSHEmulator) Init(store *Store) error {
	//	s.Config = config

	// Hardcoded context values for now; tweak later
	s.Context = NodeContext{
		CollectionName: "ssh_emulator",
		PathToContext:  "data/ssh",
		Store:          store,
	}
	log.Printf("Initialized SSHEmulator with collection: %s\n", s.Context.CollectionName)
	return nil
}

func (s *SSHEmulator) HandleInput(channel ssh.Channel) error {
	//REPL Loop
	term := terminal.NewTerminal(channel, "> ")
	chatSession := NewChatSession("ssh_connection", "ssh_emulator", s.Context.Store)

	log.Printf("channel, %v", channel)
	for {
		line, err := term.ReadLine()
		if err != nil {
			log.Printf("Failed to read line, err: %v", err)
			break
		}
		resp, err := chatSession.GenerateResponse(line)
		if err != nil {
			log.Printf("Failed to generate response, err: %v", err)
			return err
		}
		term.Write([]byte(line + "\r\n"))
		term.Write([]byte(resp + "\r\n"))
	}
	defer channel.Close()
	return nil

}
func (s *SSHEmulator) GetContext() (NodeContext, error) {
	return s.Context, nil
}
func (s *SSHEmulator) Close() error {
	return nil
}

//// REPL-based handler piping into fuxa
//func (s *SSHEmulator) HandleInput(channel ssh.Channel) error {
//	defer channel.Close()
//
//	// Setup SSH config to connect to container (update creds!)
//	config := &ssh.ClientConfig{
//		User: "root", // or whatever username your FUXA container uses
//		Auth: []ssh.AuthMethod{
//			ssh.Password("fuxaPassword"), // replace with your actual container password
//		},
//		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // DO NOT use in production
//	}
//
//	// Address of the FUXA container
//	addr := "cityucapstonemscs-fuxa-1:2222" // or "localhost:2222" if mapped
//
//	// Dial the container
//	client, err := ssh.Dial("tcp", addr, config)
//	if err != nil {
//		log.Printf("failed to connect to container: %v", err)
//		return err
//	}
//	defer client.Close()
//
//	// Create session
//	session, err := client.NewSession()
//	if err != nil {
//		log.Printf("failed to create session: %v", err)
//		return err
//	}
//	defer session.Close()
//
//	// Get stdin/stdout/stderr
//	containerIn, _ := session.StdinPipe()
//	containerOut, _ := session.StdoutPipe()
//	containerErr, _ := session.StderrPipe()
//
//	// Start a shell
//	if err := session.Shell(); err != nil {
//		log.Printf("failed to start shell: %v", err)
//		return err
//	}
//
//	// Proxy data
//	go io.Copy(containerIn, channel)           // user input → container
//	go io.Copy(channel, containerOut)          // container output → user
//	go io.Copy(channel.Stderr(), containerErr) // container stderr → user
//
//	// Wait for session to end
//	if err := session.Wait(); err != nil {
//		log.Printf("session finished with error: %v", err)
//	}
//
//	return nil
//}
