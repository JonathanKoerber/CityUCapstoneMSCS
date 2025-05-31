package emulator

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
