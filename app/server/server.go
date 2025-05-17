package server

type ProtocolTCPServer interface {
	Start()
	Reset()
	Stop()
}
