package emulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

type Emulator interface {
	// New create a new emulator with node-specific data
	Init(config NodeConfig) error

	// HandleCommand Handles input to protocol server
	HandleCommand(sessionID string, input string) (string, error)

	GetContext() (NodeContext, error)

	// Close releases any resource used by emulator
	Close() error
}

type QdrantStore struct {
	Client      qdrant.Client
	VectorDBURI string
	Port        int
	OllamaURI   string
	Model       string
	Collections []string
}

type NodeConfig struct {
	Port         int
	Protocol     string
	Role         string
	Service      string
	ServiceName  string
	SystemPrompt string
	Tags         []string
	User         string
}

type NodeContext struct {
	CollectionName         string
	PathToContext          string
	Distance               qdrant.Distance
	DefaultSegmentDistance qdrant.Distance
	DefaultSegmentNumber   *uint64
	VectorSize             uint64
	Store                  QdrantStore
}

func (store *QdrantStore) Init(emulators []Emulator) error {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: store.VectorDBURI,
		Port: store.Port,
		// APIKey: "<API_KEY>",
		// UseTLS: true,
		// TLSConfig: &tls.Config{},
		// GrpcOptions: []grpc.DialOption{},
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	healthCheckResult, err := client.HealthCheck(ctx)
	if err != nil {
		log.Fatalf("Could not connect to qdrant %v: %v", store.VectorDBURI, err)
	}
	log.Printf("QdrantStore Health Check Result: %v", healthCheckResult)
	for _, emulator := range emulators {
		nodeContext, err := emulator.GetContext()
		exists, err := client.CollectionExists(ctx, nodeContext.CollectionName)
		if err != nil {
			log.Fatalf("Could not check if collection exists: %v", err)
			return err
		}
		// check to see if the collection name exists
		if !exists {
			err = client.CreateCollection(ctx, &qdrant.CreateCollection{
				CollectionName: nodeContext.CollectionName,
				VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
					Size:     nodeContext.VectorSize,
					Distance: nodeContext.Distance,
				}),
				OptimizersConfig: &qdrant.OptimizersConfigDiff{
					DefaultSegmentNumber: nodeContext.DefaultSegmentNumber,
				},
			})
			if err != nil {
				log.Fatalf("Could not create collection: %v", err)
			} else {
				log.Printf("Created collection: %v", nodeContext.CollectionName)
			}
		}
	} // > error here exception something
	collection, err := client.ListCollections(ctx)
	store.Collections = collection
	if err != nil {
		log.Fatalf("Could not list collections: %v", err)
	} else {
		log.Printf("Listing collections: %s", &collection)
	}
	return nil
}

func (store *QdrantStore) generateEmbedding(prompt string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model":  store.Model,
		"prompt": prompt,
	}
	reqBytes, err := json.Marshal(reqBody)
	resp, err := http.Post(store.OllamaURI+"/api/embeddings", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("Embedding request failded with %v", err)
	}
	log.Printf("emulator response: %v", resp)
	defer resp.Body.Close()
	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Failded to decode embedding: %w", err)
	}
	return result.Embedding, nil
}

func (store *QdrantStore) VectorContext(emulator Emulator) ([]*qdrant.PointStruct, error) {
	nodeContext, err := emulator.GetContext()
	if err != nil {
		log.Fatalf("Could not get context for %v: %v", nodeContext.CollectionName, err)
	}
	//  TODO walk dir and vectorize files.
	// TODO this is a dir need to embed alll the files that are in thid the dir.
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
			log.Fatalf("Could not read file: %v", err)
		}
		lines = strings.Split(string(data), "\n")
	}
	var result []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	var points []*qdrant.PointStruct
	for _, line := range result {
		vector, err := store.generateEmbedding(line) // you implement this
		if err != nil {
			return nil, err
		}
		point := qdrant.PointStruct{
			Vectors: qdrant.NewVectors(vector...),
			Payload: map[string]*qdrant.Value{
				"command": qdrant.NewValueString(line),
			},
		}
		points = append(points, &point)
	}
	return points, nil
}

func (store *QdrantStore) AddVectors(collectionName string, vectors []*qdrant.PointStruct) error {
	ctx := context.Background()
	_, err := store.Client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         vectors,
	})
	return err
}
