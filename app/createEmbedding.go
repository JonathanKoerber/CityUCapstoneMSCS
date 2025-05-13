package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
)

func CreateOllamaEmbedding(text string) ([]float32, error) {
	model := os.Getenv("OlamaModel")
	if model == "" {
		return nil, fmt.Errorf("Environment variable OLAMA_MODEL not set")
	}
	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": text,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post("http://ollama:11434/api/embeddings", "application/json", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Embeddings []float32 `json:"embeddings"`
	}
	// assigns decode json to result struct
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Embeddings, nil
}

func EnbedContextFiles() {
	basePath := os.Getenv("CONTEXT_DATA_BASE_PATH")
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != basePath {
			contextName := filepath.Base(path)
			collectionName := fmt.Sprintf("honeypot_%s", contextName)

		}
	})
}
