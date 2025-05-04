package server

type ProtocolTCPServer interface {
	Start(port int)
	Reset()
	Stop()
}
