package server

type ProtocolServer interface {
	Start()
	Reset()
	Stop()
}
