package temporal

// Repo defines APIs for Temporal to access persistence layer.
type Repo interface {
	SetNodeAt(name []byte, height int32) error
	NodesAt(height int32) ([][]byte, error)
	Close() error
}
