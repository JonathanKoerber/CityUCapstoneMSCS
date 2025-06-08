package emulator

import (
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"time"
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

//	func (s *SSHEmulator) HandleInput(channel ssh.Channel) error {
//		//REPL Loop
//		term := terminal.NewTerminal(channel, "> ")
//		chatSession := NewChatSession("ssh_connection", "ssh_emulator", s.Context.Store)
//
//		log.Printf("channel, %v", channel)
//		for {
//			line, err := term.ReadLine()
//			if err != nil {
//				log.Printf("Failed to read line, err: %v", err)
//				break
//			}
//			resp, err := chatSession.GenerateResponse(line)
//			if err != nil {
//				log.Printf("Failed to generate response, err: %v", err)
//				return err
//			}
//			term.Write([]byte(line + "\r\n"))
//			term.Write([]byte(resp + "\r\n"))
//		}
//		defer channel.Close()
//		return nil
//
// }
func (s *SSHEmulator) GetContext() (NodeContext, error) {
	return s.Context, nil
}
func (s *SSHEmulator) Close() error {
	return nil
}

// REPL-based handler piping into fuxa
func (s *SSHEmulator) HandleInput(channel ssh.Channel) error {
	defer channel.Close()

	config := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			ssh.Password("password"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	addr := "ics-host:22" // Make sure this matches the container SSH port

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Printf("failed to connect to container: %v", err)
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Printf("failed to create session: %v", err)
		return err
	}
	defer session.Close()

	// Request PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Printf("failed to request PTY: %v", err)
		return err
	}

	// Attach I/O
	containerIn, _ := session.StdinPipe()
	containerOut, _ := session.StdoutPipe()
	containerErr, _ := session.StderrPipe()

	if err := session.Shell(); err != nil {
		log.Printf("failed to start shell: %v", err)
		return err
	}

	go io.Copy(containerIn, channel)
	go io.Copy(channel, containerOut)
	go io.Copy(channel.Stderr(), containerErr)

	if err := session.Wait(); err != nil {
		log.Printf("session finished with error: %v", err)
	}

	return nil
}
