package temporal

type Repo interface {
	SetNodeAt(name string, height int32) error
	NodesAt(height int32) ([]string, error)
	Close() error
}
