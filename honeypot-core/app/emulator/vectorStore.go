package emulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
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
	Client      *qdrant.Client
	Collections []string
}

func NewEmulator() Store {
	return Store{}

}

func (store *Store) Init() error {
	//vecStoreUri := os.Getenv("VEC_STORE_URI")
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "qdrant",
		Port: 6334,
	})
	if err != nil {
		log.Fatalf("Fatal error connecting to VEC store: %v", err)
	}
	// add qdrant client to vector store
	store.Client = client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	healthCheckResult, err := client.HealthCheck(ctx)
	if err != nil {
		log.Fatalf("Fatal error connecting to VEC store: %v", err)
	}
	log.Printf("Health check result: %v", healthCheckResult)
	return err
}

func (store *Store) ReadContextFiles(nodeContext NodeContext) ([]string, error) {
	//  TODO walk dir and vectorized files.
	// TODO this is a dir need to embed all the files that are in third the dir.
	data, err := os.ReadDir(nodeContext.PathToContext)
	if err != nil {
		log.Fatalf("Could not read context: %v", err)
	}
	var lines []string
	for _, entry := range data {
		if entry.IsDir() {
			//dirPath := filepath.Join(nodeContext.PathToContext, entry.ImageName())
			//data.append(data, os.ReadDir(dirPath))
			continue
		}
		filePath := filepath.Join(nodeContext.PathToContext, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Could not read file: %v", err)
		}
		lines = strings.Split(string(data), "\n")
		for line := range lines {
			lines[line] = strings.TrimSpace(lines[line])
		}
	}
	return lines, nil
}

func (store *Store) EmbedDocs(line string) (*qdrant.PointStruct, error) {
	id := uuid.New().String()
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
	//embedding := make([]float64, len(result.Embedding))
	payload := map[string]interface{}{"content": line}
	if err != nil {
		return nil, fmt.Errorf("failed to create payload: %w", err)
	}
	if len(result.Embedding) == 0 {
		log.Printf("no embedding found: %s", line)
		return nil, nil
	}
	point := &qdrant.PointStruct{
		Id: &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{Uuid: id}, // or use UIntId if numeric
		},
		Vectors: qdrant.NewVectors(result.Embedding...),
		Payload: qdrant.NewValueMap(payload),
	}

	return point, nil
}

func (store *Store) CreateCollection(collectionName string, vectorSize uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var defaultSeg *uint64
	v := uint64(2)
	defaultSeg = &v
	err := store.Client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Dot,
		}),
		OptimizersConfig: &qdrant.OptimizersConfigDiff{
			DefaultSegmentNumber: defaultSeg,
		},
	})
	if err != nil {
		log.Fatalf("Fatal error creating collection: %v", err)
	}
	return err
}

func (store *Store) AddVectors(collectionName string, upsertPoints []*qdrant.PointStruct) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connect, err := store.Client.CollectionExists(ctx, collectionName)
	if err != nil {
		log.Fatalf("Fatal error connecting to collection: %v", err)
	}
	if !connect {
		store.CreateCollection(collectionName, 2560)
	}
	_, err = store.Client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         upsertPoints,
	})
	if err != nil {
		log.Fatalf("Could not upsert points: %v", err)
	}
	log.Println("Upsert", len(upsertPoints), "points")
	return err
}
