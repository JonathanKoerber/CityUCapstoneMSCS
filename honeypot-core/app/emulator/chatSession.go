package emulator

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/qdrant/go-client/qdrant"
	"log"
	"net/http"
	"os"
	"strings"
)

type ChatSession struct {
	CollectionName string
	History        []string
	Context        []int
	Store          *Store
	SystemID       string
}

type RetrievedDocs struct {
	Content string  `bson:"content"`
	Score   float64 `bson:"score,omitempty"`
}

func NewChatSession(clientID string, collectionName string, store *Store) *ChatSession {
	return &ChatSession{
		History:        make([]string, 0),
		Store:          store,
		SystemID:       clientID,
		CollectionName: collectionName,
	}
}

func (session *ChatSession) GetTopKVectors(queryPoint *qdrant.PointStruct, topk int) ([]*qdrant.PointGroup, error) {
	ctx := context.Background()
	vectors := queryPoint.Vectors.GetVector().Data
	resp, err := session.Store.Client.QueryGroups(ctx, &qdrant.QueryPointGroups{
		CollectionName: session.CollectionName,
		Query:          qdrant.NewQuery(vectors...),
		GroupBy:        "document_id",
		GroupSize:      qdrant.PtrOf(uint64(topk)),
	})
	if err != nil {
		log.Printf("QueryPointGroups: %v", err)
		return nil, err
	}

	return resp, nil
}

func (session *ChatSession) GetLLMResponse(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":  os.Getenv("MODEL"),
		"prompt": prompt,
		"stream": false,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(os.Getenv("OLLAMA_URL")+"/api/generate", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var fullResponse strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue // skip empty lines
		}

		var chunk struct {
			Response   string `json:"response"`
			Done       bool   `json:"done"`
			Context    []int  `json:"context"`
			DoneReason string `json:"done_reason"`
		}

		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			// log error but keep processing next chunks if possible
			log.Printf("Failed to unmarshal chunk: %v, line: %s", err, line)
			continue
		}

		// Update session context if present
		if len(chunk.Context) > 0 {
			session.Context = chunk.Context
		}

		// Append the partial response text
		fullResponse.WriteString(chunk.Response)

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return fullResponse.String(), nil
}

func (session *ChatSession) GenerateResponse(userInput string) (string, error) {
	// This links the whole process to geather
	embeddingInput, err := session.Store.EmbedDocs(userInput)
	if err != nil {
		log.Printf("Error creating embedding input: %v", err)
		return "", err
	}
	pointGroup, err := session.GetTopKVectors(embeddingInput, 3)
	if err != nil {
		log.Printf("Error getting embedding input: %v", err)
	}
	contextText := ""
	for _, retrievedDoc := range pointGroup {
		contextText += retrievedDoc.Lookup.String() + "\n"
	}
	prompt := fmt.Sprintf("Context:\n%s\n\nUser: %s\n\nAssistant: %s\n\n", contextText, userInput)
	response, err := session.GetLLMResponse(prompt)
	if err != nil {
		log.Printf("Error creating prompt: %v", err)
	}
	return response, nil
}
