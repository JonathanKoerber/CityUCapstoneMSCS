package emulator

import (
	"github.com/qdrant/go-client/qdrant"
	"log"
	"time"
)

type NodeContext struct {
	CollectionName string
	PathToContext  string
	Store          *Store
}

type EmbeddedDocs struct {
	Embeddings []float64 `bson:"embeddings"`
	Line       string    `bson:"line"`
}

type Emulator interface {
	Init()
}

func EmbedContext(context NodeContext, store Store) error {

	lineChan := make(chan string, 100)
	embeddingChan := make(chan *qdrant.PointStruct, 1000)
	log.Printf("Starting to embed data")
	go func() {
		defer close(lineChan)
		lines, err := store.ReadContextFiles(context)
		if err != nil {
			log.Printf("Failed to read context files: %v", err)
		}
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			select {
			case lineChan <- line:
			default:
				log.Printf("Line channel full, waiting...")
				time.Sleep(1 * time.Second)
				lineChan <- line
			}
		}
	}()
	go func() {
		defer close(embeddingChan)
		for line := range lineChan {
			embedding, err := store.EmbedDocs(line)
			if err != nil {
				log.Printf("Failed to embed context: %v", err)
				continue
			}
			select {
			case embeddingChan <- embedding:
			default:
				log.Printf("Embedding channel full, waiting...")
				time.Sleep(1 * time.Second)
			}
		}
	}()
	go func() {
		batch := make([]*qdrant.PointStruct, 0, 100)
		ticker := time.NewTicker(2 * time.Second)
		batchSize := 20
		defer ticker.Stop()
		for {
			select {
			case embedding, ok := <-embeddingChan:
				if !ok {
					if len(batch) > 0 {
						store.AddVectors("ssh_emulator", batch)
					}
					log.Printf("Ebedding err: %v", ok)
					return
				}
				if embedding != nil {
					batch = append(batch, embedding)
				}
				batch = append(batch, embedding)
				if len(batch) > batchSize {
					store.AddVectors("ssh_emulator", batch)
					batch = batch[:0]
				}
			case <-ticker.C:
				if len(batch) > 0 {
					store.AddVectors("ssh_emulator", batch)
					batch = batch[:0]
				}
			}
		}
	}()
	return nil
}
