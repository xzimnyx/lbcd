package temporalrepo

type Memory struct {
	cache map[int32]map[string]bool
}

func NewMemory() *Memory {
	return &Memory{
		cache: map[int32]map[string]bool{},
	}
}

func (repo *Memory) SetNodeAt(name []byte, height int32) error {

	names, ok := repo.cache[height]
	if !ok {
		names = map[string]bool{}
		repo.cache[height] = names
	}
	names[string(name)] = true

	return nil
}

func (repo *Memory) NodesAt(height int32) ([][]byte, error) {

	var names [][]byte

	for name := range repo.cache[height] {
		names = append(names, []byte(name))
	}

	return names, nil
}

func (repo *Memory) Close() error {
	return nil
}
