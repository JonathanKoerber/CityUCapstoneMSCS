package emulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
)

type ChatSession struct {
	History  []string
	Store    *Store
	SystemID string
}

type RetrievedDocs struct {
	Content string  `bson:"content"`
	Score   float64 `bson:"score,omitempty"`
}

func NewChatSession(clientID string, store *Store) *ChatSession {
	return &ChatSession{History: make([]string, 0), Store: store, SystemID: clientID}
}

func (session *ChatSession) GetTopKVectors(embeddedInput []float32, topk int) ([]RetrievedDocs, error) {
	collection := session.Store.Client.Database("honeypot").Collection("ssh_context")
	pipeline := mongo.Pipeline{
		bson.D{
			{"$search", bson.D{
				{"index", "vector_index"},
				{"knnBeta", bson.D{
					{"vector", embeddedInput},
					{"path", "embedding"},
					{"k", topk},
				}},
			}},
		},
		bson.D{{"$project", bson.D{
			{"content", 1},
			{"score", bson.D{{"$meta", "searchScore"}}},
		}}},
	}
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	var results []RetrievedDocs
	if err := cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (session *ChatSession) GetLLMResponse(prompt string) (string, error) {
	regBody := map[string]interface{}{
		"model":  os.Getenv("MODEL"),
		"prompt": prompt,
	}
	recBytes, err := bson.Marshal(regBody)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(os.Getenv("OLLAMA_URL")+"api/completions", "application/json", bytes.NewBuffer(recBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var res struct {
		Completion string `json:"completion"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.Completion, nil
}

func (session *ChatSession) GenerateResponse(userInput string) (string, error) {
	embeddingInput, err := session.Store.EmbedString(userInput)
	if err != nil {
		log.Printf("Error creating embedding input: %v", err)
		return "", err
	}
	retrievedDocs, err := session.GetTopKVectors(embeddingInput, 3)
	if err != nil {
		log.Printf("Error getting embedding input: %v", err)
	}
	contextText := ""
	for _, retrievedDoc := range retrievedDocs {
		contextText += retrievedDoc.Content + "\n"
	}
	prompt := fmt.Sprintf("Context:\n%s\n\nUser: %s\n\nAssistant: %s\n\n", contextText, userInput)
	response, err := session.GetLLMResponse(prompt)
	if err != nil {
		log.Printf("Error creating prompt: %v", err)
	}
	session.History = append(session.History, response)
	return response, nil
}
