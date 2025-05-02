package server

import "golang.org/x/crypto/ssh"

type ProtocolServer interface {
	Start(*ssh.ServerConfig, int) error
	Reset()
	Stop()
}
