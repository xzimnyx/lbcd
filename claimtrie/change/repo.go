package change

// TODO: swtich to steam-like read API, such as iterartor, or scanner.
type ChainChangeRepo interface {
	Save(changes []Change) error
	LoadByHeight(height int32) ([]Change, error)
	Close() error
}

type NodeChangeRepo interface {
	Save(changes []Change) error
	LoadByNameUpToHeight(name string, height int32) ([]Change, error)
	Close() error
}
