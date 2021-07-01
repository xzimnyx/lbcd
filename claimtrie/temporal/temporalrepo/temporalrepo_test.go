package temporalrepo

import (
	"testing"

	"github.com/btcsuite/btcd/claimtrie/temporal"
	"github.com/stretchr/testify/assert"
)

func TestMemory(t *testing.T) {

	repo := NewMemory()
	testTemporalRepo(t, repo)
}

func TestPebble(t *testing.T) {

	repo, err := NewPebble(t.TempDir())
	if !assert.NoError(t, err) {
		return
	}

	testTemporalRepo(t, repo)
}

func testTemporalRepo(t *testing.T, repo temporal.Repo) {

	nameA := []byte("a")
	nameB := []byte("b")
	nameC := []byte("c")

	testcases := []struct {
		name    []byte
		heights []int32
	}{
		{nameA, []int32{1, 3, 2}},
		{nameA, []int32{2, 3}},
		{nameB, []int32{5, 4}},
		{nameB, []int32{5, 1}},
		{nameC, []int32{4, 3, 8}},
	}

	for _, i := range testcases {
		for _, height := range i.heights {
			err := repo.SetNodeAt(i.name, height)
			assert.NoError(t, err)
		}
	}

	// a: 1, 2, 3
	// b: 1, 5, 4
	// c: 4, 3, 8

	names, err := repo.NodesAt(2)
	assert.NoError(t, err)
	assert.ElementsMatch(t, [][]byte{nameA}, names)

	names, err = repo.NodesAt(5)
	assert.NoError(t, err)
	assert.ElementsMatch(t, [][]byte{nameB}, names)

	names, err = repo.NodesAt(8)
	assert.NoError(t, err)
	assert.ElementsMatch(t, [][]byte{nameC}, names)

	names, err = repo.NodesAt(1)
	assert.NoError(t, err)
	assert.ElementsMatch(t, [][]byte{nameA, nameB}, names)

	names, err = repo.NodesAt(4)
	assert.NoError(t, err)
	assert.ElementsMatch(t, [][]byte{nameB, nameC}, names)

	names, err = repo.NodesAt(3)
	assert.NoError(t, err)
	assert.ElementsMatch(t, [][]byte{nameA, nameC}, names)
}
