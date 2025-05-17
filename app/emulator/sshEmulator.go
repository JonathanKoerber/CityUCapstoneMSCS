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
		CollectionName: "ssh_context",
		PathToContext:  "data/ssh",
		Store:          store,
	}
	log.Printf("Initialized SSHEmulator with collection: %s\n", s.Context.CollectionName)
	return nil
}

func (s *SSHEmulator) HandleInput(channel ssh.Channel) error {
	//REPL Loop
	term := terminal.NewTerminal(channel, "> ")
	chatSession := NewChatSession("ssh_connection", s.Context.Store)

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
