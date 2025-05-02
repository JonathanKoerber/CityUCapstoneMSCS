package server

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func Listen(port int) {

	authorizedKeysBytes, err := os.ReadFile("auth_keys/authorized_keys")
	if err != nil {
		log.Fatal("Failed to load auth keys, err: %v", err)
	}
	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Fatal("Failed to parse authorized_keys, err: %v", err)
		}
		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// this  login with password
			if c.User() == "root" && string(pass) == "pass" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		// public key auth
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}
	privateBytes, err := os.ReadFile("ssh_keys/id_rsa")
	if err != nil {
		log.Fatal("Failed to load private keys, err: %v", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private keys, err: %v", err)
	}
	config.AddHostKey(private)

	// host config done host can now be configured
	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Fatal("Failed to listen, err: %v", err)
	}

	// create network connection
	nConn, err := listener.Accept()
	if err != nil {
		log.Fatal("Failed to accept incoming connections, err: %v", err)
	}
	log.Printf("Accepted incoming connection from %s", nConn.RemoteAddr())

	// befor conn used
	// handshake must be preformed on the incomming conn
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Fatal("Failed to handshake, err: %v", err)
	}
	log.Printf("New SSH connection from %s", conn.RemoteAddr())

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
	}()

	// loop to handle multible requests
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatal("Failed to accept channel, err: %v", err)
		}
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		wg.Add(1)
		go func(in <-chan *ssh.Request) {
			for req := range in {
				req.Reply(req.Type == "shell", nil)
			}
			wg.Done()
		}(requests)
		term := terminal.NewTerminal(channel, "> ")

		wg.Add(1)
		go func() {
			defer func() {
				channel.Close()
				wg.Done()
			}()
			for {
				line, err := term.ReadLine()
				if err != nil {
					break
				}
				fmt.Println(line)
			}
		}()
	}

}

func (s *SSHServer) Start(config *ssh.ServerConfig, port int) error {
	listener, err := net.Listen("tcp", "0.0.0.0"+com.ToStr(port))
	for {
		// when server has been congiured , conn can be added
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		// before uses a handshake must be configured
		sConn, chans, reqs, err := ssh.NewServerConn(conn, config)
		if err != nil {
			continue
		}
		go ssh.DiscardRequest(reqs)

	}
}

func (s *SSHServer) Reset() {}

func (s *SSHServer) Stop() error {
	if s.listener != nil {
		fmt.Printf("[-] SSH server stopping\n")
		return s.listener.Close()
	}
	return nil
}
