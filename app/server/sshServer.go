package server

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"net"
)

type SSHServer struct {
	Addr     string
	Server   *ssh.Server
	listener net.Listener
}

func NewSSHServer(addr string) *SSHServer {
	return &SSHServer{Addr: addr}
}

func (s *SSHServer) Start() error {
	s.Server = &ssh.Server{
		Addr: s.Addr,
		Handler: func(sess ssh.Session) {
			io.WriteString(sess, "Hi Welcome to SCADA honeypot!\n")
			command := sess.RawCommand()
			fmt.Printf("[SSH] %s rang: %s\n", sess.RemoteAddr().String(), command)
		},
	}
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("listen failed: %v", err)
	}
	s.listener = l
	go s.Server.Serve(l)

	fmt.Printf("[+] SSH server listening on %s\n", s.Addr)
	return nil
}

func (s *SSHServer) Reset() {}

func (s *SSHServer) Stop() error {
	if s.listener != nil {
		fmt.Printf("[-] SSH server stopping\n")
		return s.listener.Close()
	}
	return nil
}
