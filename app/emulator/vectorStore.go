package emulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VectorStore interface {
	// New create a new emulator with node-specific data
	Init(store *Store) error

	// HandleCommand Handles input to protocol server
	HandleCommand(sessionID string, input string) (string, error)

	GetContext() (NodeContext, error)

	// Close releases any resource used by emulator
	Close() error
}

type Store struct {
	Client      *mongo.Client
	Collections []string
}

func NewEmulator() Store {
	return Store{}

}

func (store *Store) Init() error {
	mangoURI := os.Getenv("VEC_STORE_URI")
	log.Printf("VEC_STORE_URI: %s", mangoURI)
	clientOptions := options.Client().ApplyURI(mangoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Printf("Error initializing MONGODB is not connecting: %v", err)
		cancel()
		return err
	}
	store.Client = client
	defer cancel()
	return nil
}

func (store *Store) ReadContextFiles(nodeContext NodeContext) ([]string, error) {
	//  TODO walk dir and vectorize files.
	// TODO this is a dir need to embed all the files that are in third the dir.
	data, err := os.ReadDir(nodeContext.PathToContext)
	if err != nil {
		log.Fatalf("Could not read context: %v", err)
	}
	var lines []string
	for _, entry := range data {
		if entry.IsDir() {
			//dirPath := filepath.Join(nodeContext.PathToContext, entry.Name())
			//data.append(data, os.ReadDir(dirPath))
			continue
		}
		filePath := filepath.Join(nodeContext.PathToContext, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Could not read file: %v", err)
		}
		lines = strings.Split(string(data), "\n")
	}
	return lines, nil
}
func (store *Store) EmbedString(line string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model":  os.Getenv("MODEL"),
		"prompt": line,
	}
	reqBytes, err := json.Marshal(reqBody)
	resp, err := http.Post(os.Getenv("OLLAMA_URL")+"/api/embeddings", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("embedding request failded with: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding request failded with: %v - %v", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failded to decode embedding: %w", err)
	}
	return result.Embedding, nil
}
func (store *Store) EmbedDocs(line string) (*EmbeddedDocs, error) {
	reqBody := map[string]interface{}{
		"model":  os.Getenv("MODEL"),
		"prompt": line,
	}
	reqBytes, err := json.Marshal(reqBody)
	resp, err := http.Post(os.Getenv("OLLAMA_URL")+"/api/embeddings", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("embedding request failded with: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding request failded with: %v - %v", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failded to decode embedding: %w", err)
	}
	embedding := make([]float64, len(result.Embedding))

	return &EmbeddedDocs{embedding, line}, nil
}

func (store *Store) AddVectors(collectionName string, vector *EmbeddedDocs) error {

	collection := store.Client.Database("honeypot").Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	resp, err := collection.InsertOne(ctx, vector)
	if err != nil {
		log.Printf("Could not insert mongo honeypot database: %v", err)
	}
	log.Printf("Inserted mongo %v %v", collectionName, resp)
	defer cancel()
	return err
}
