package temporalrepo

type Memory struct {
	cache map[int32]map[string]bool
}

func NewMemory() *Memory {
	return &Memory{
		cache: map[int32]map[string]bool{},
	}
}

func (repo *Memory) NodesAt(height int32) ([]string, error) {

	var names []string

	for name := range repo.cache[height] {
		names = append(names, name)
	}

	return names, nil
}

func (repo *Memory) SetNodeAt(name string, height int32) error {

	names, ok := repo.cache[height]
	if !ok {
		names = map[string]bool{}
		repo.cache[height] = names
	}
	names[name] = true

	return nil
}

func (repo *Memory) Close() error {
	return nil
}
