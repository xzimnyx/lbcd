package repo

type TemporalRepoMem struct {
	cache map[int32]map[string]bool
}

func NewTemporalMem() *TemporalRepoMem {
	return &TemporalRepoMem{
		cache: map[int32]map[string]bool{},
	}
}

func (repo *TemporalRepoMem) NodesAt(height int32) ([]string, error) {

	var names []string

	for name := range repo.cache[height] {
		names = append(names, name)
	}

	return names, nil
}

func (repo *TemporalRepoMem) SetNodeAt(name string, height int32) error {

	names, ok := repo.cache[height]
	if !ok {
		names = map[string]bool{}
		repo.cache[height] = names
	}
	names[name] = true

	return nil
}

func (repo *TemporalRepoMem) Close() error {
	return nil
}
