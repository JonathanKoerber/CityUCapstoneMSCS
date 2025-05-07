package emulator

type Emulator interface {
	// New create a new emulator with node-specific data
	New(config NodeConfig) error

	// HandleCommand Handles input to protocol server
	HandleCommand(sessionID string, input string) (string, error)

	// Close releases any resource used by emulator
	Close() error
}

type NodeConfig struct {
	NodeID       string
	Role         string
	VectorDBURI  string
	Modle        string
	SystemPrompt string
}
