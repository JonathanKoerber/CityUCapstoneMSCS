// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/crypto/blob/master/ssh/example_test.go

package server

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"net"
	"os"
	"sync"
)

type SSHServer struct {
	Config *ssh.ServerConfig
	Port   int
	ln     net.Listener
}

func NewSSHServer() (*SSHServer, error) {
	return &SSHServer{}, nil
}

func (s *SSHServer) Start(port int) {
	// Todo: Figure how I want to auth.
	log.Printf("Starting server on port %d\n", port)
	authorizedKeysBytes, err := os.ReadFile("authorized_keys")

	if err != nil {
		log.Printf("Failed to load auth keys, err: %v", err)
	}
	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Printf("Failed to parse authorized_keys, err: %v", err)
		}
		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	s.Config = &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// this  login with password
			if c.User() == "admin" && string(pass) == "password" {
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
		log.Printf("Failed to load private keys, err: %v", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Printf("Failed to parse private keys, err: %v", err)
	}
	s.Config.AddHostKey(private)

	// host config done host can now be configured
	s.ln, err = net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Printf("Failed to listen, err: %v", err)
	}

	// create network connection
	nConn, err := s.ln.Accept()
	if err != nil {
		log.Printf("Failed to accept incoming connections, err: %v", err)
	}
	log.Printf("Accepted incoming connection from %s", nConn.RemoteAddr())

	// befor conn used
	// handshake must be preformed on the incomming conn
	conn, chans, reqs, err := ssh.NewServerConn(nConn, s.Config)
	if err != nil {
		log.Printf("Failed to handshake, err: %v", err)
	}
	if conn.Permissions != nil {
		if fp, ok := conn.Permissions.Extensions["pubkey-fp"]; ok {
			log.Printf("New SSH connection from %s", fp)
		} else {
			log.Printf("New SSH connection (no fingerprint)")
		}
	} else {
		log.Printf("New SSH connection (no permissions metadata)")
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(2)
	go func(in <-chan *ssh.Request) {
		for req := range in {
			switch req.Type {
			case "pty-req":
				req.Reply(true, nil)
			case "shell":
				req.Reply(true, nil)
			default:
				req.Reply(false, nil)
			}
		}
		wg.Done()
	}(reqs)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
	}()

	// loop to handle multiple requests spins up async process for each request
	// Todo: Handle DDoS here?
	for newChannel := range chans {
		log.Printf("New channel from %s", newChannel.ChannelType(), newChannel.ExtraData())
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
			log.Printf("Failed to accept channel, err: %v", err)
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
		// Todo: pip to system that will be attacked.
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

// TODO add to the SSHServer struct
func ServerConfig_AddHostKey() {
	// only password auth
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Todo: handle user name and password
			if c.User() == "admin" && string(pass) == "password" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}
	privateBytes, err := os.ReadFile("ssh_keys/id_rsa")
	if err != nil {
		log.Printf("Failed to load private keys, err: %v", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Printf("Failed to parse private keys, err: %v", err)
	}

	// Restrict host key algorithms to disable ssh-rsa
	signer, err := ssh.NewSignerWithAlgorithms(private.(ssh.AlgorithmSigner), []string{ssh.KeyAlgoECDSA256, ssh.KeyAlgoRSASHA512})
	if err != nil {
		log.Printf("Failed to create private key with restricted algorithms: %v", err)
	}
	config.AddHostKey(signer)
}

func (s *SSHServer) Reset() {
	if s.ln != nil {
		s.ln.Close()
	}
	s.Start(s.Port)
}

func (s *SSHServer) Stop() {
	if s.ln != nil {
		s.ln.Close()
	}
}
