package emulator

import (
	"log"
)

type SSHEmulator struct {
	//Config  NodeConfig
	Context NodeContext
}

func NewSSHEmulator() *SSHEmulator {
	return &SSHEmulator{}
}
func (s *SSHEmulator) Init(store *EmbeddingStore) error {
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

func (s *SSHEmulator) HandleCommand(sessionID string, input string) (string, error) {
	response := "hello world"
	return response, nil

}
func (s *SSHEmulator) GetContext() (NodeContext, error) {
	return s.Context, nil
}
func (s *SSHEmulator) Close() error {
	return nil
}
