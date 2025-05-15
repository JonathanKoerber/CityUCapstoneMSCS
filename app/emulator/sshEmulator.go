package emulator

import (
	"github.com/qdrant/go-client/qdrant"
	"log"
)

type SSHEmulator struct {
	//Config  NodeConfig
	Context NodeContext
}

func NewSSHEmulator() *SSHEmulator {
	return &SSHEmulator{}
}
func (s *SSHEmulator) Init(store *QdrantStore) error {
	//	s.Config = config

	// Hardcoded context values for now; tweak later
	s.Context = NodeContext{
		CollectionName:         "ssh_context",
		PathToContext:          "data/ssh",
		Distance:               qdrant.Distance_Cosine,
		DefaultSegmentDistance: qdrant.Distance_Cosine,
		VectorSize:             1536, // typical OpenAI/embedding vector size
		DefaultSegmentNumber:   nil,  // or a pointer if needed
		Store:                  store,
	}
	log.Printf("Initialized SSHEmulator with collection: %s\n", s.Context.CollectionName)
	return nil
}

func (s *SSHEmulator) HandleCommand(sessionID string, input string) (string, error) {

	//userPrompt := `Create a "hello, world" programing in GO`
	//options := llm.SetOptions(map[string]interface{}{
	//	"temperature":    0.5,
	//	"repeat_last_n":  2,
	//	"repeat_penalty": 2.2,
	//})
	//query := llm.Query{
	//	Model: s.config.Model,
	//	Messages: []llm.Message{
	//		{Role: "system", Content: s.config.SystemPrompt},
	//		{Role: "user", Content: userPrompt},
	//	},
	//	Options: options,
	//}
	//response := query.ToJsonString()
	response := "hello world"
	return response, nil

}
func (s *SSHEmulator) GetContext() (NodeContext, error) {
	return s.Context, nil
}
func (s *SSHEmulator) Close() error {
	return nil
}
