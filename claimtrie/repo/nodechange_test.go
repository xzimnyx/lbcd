package repo

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/btcsuite/btcd/claimtrie/change"

	"github.com/stretchr/testify/assert"
)

var (
	opStr1        = "0000000000000000000000000000000000000000000000000000000000000000:1"
	testNodeName1 = []byte("name1")
)

func TestNodeChangeRepoPebble(t *testing.T) {

	file, err := ioutil.TempFile("/tmp", "claimtrie_test")
	if err != nil {
		log.Fatal(err)
	}
	os.Remove(file.Name())

	repo, err := NewNodeChangeRepoPebble(file.Name())
	if assert.NoError(t, err) {

		cleanup := func() {
			upperbound := append(testNodeName1, byte(1))
			repo.db.DeleteRange(testNodeName1, upperbound, nil)
		}

		testNodeChangeRepo(t, repo, func() {}, cleanup)
	}
}

func TestNodeChangeRepoPostgres(t *testing.T) {

	// repo, err := NewNodeChangeRepoPostgres(cfg.TestPostgresDB.DSN, cfg.TestPostgresDB.Drop)
	// if assert.NoError(t, err) {
	// 	testNodeChangeRepo(t, repo, func() {}, func() {})
	// }
}

func testNodeChangeRepo(t *testing.T, repo change.NodeChangeRepo, setup, cleanup func()) {

	chg := change.New(change.AddClaim).SetName(testNodeName1).SetOutPoint(opStr1)

	testcases := []struct {
		name     string
		height   int32
		changes  []change.Change
		expected []change.Change
	}{
		{
			"test 1",
			1,
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
			[]change.Change{chg.SetHeight(1)},
		},
		{
			"test 2",
			2,
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
			[]change.Change{chg.SetHeight(1)},
		},
		{
			"test 3",
			3,
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3)},
		},
		{
			"test 4",
			4,
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3)},
		},
		{
			"test 5",
			5,
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
		},
		{
			"test 6",
			6,
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
			[]change.Change{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
		},
	}

	for _, tt := range testcases {

		setup()

		err := repo.Save(tt.changes)
		assert.NoError(t, err)

		changes, err := repo.LoadByNameUpToHeight(string(testNodeName1), tt.height)
		assert.NoError(t, err)
		assert.Equalf(t, tt.expected, changes, tt.name)

		cleanup()
	}

	testcases2 := []struct {
		name     string
		height   int32
		changes  [][]change.Change
		expected []change.Change
	}{
		{
			"Save in 2 batches, and load up to 1",
			1,
			[][]change.Change{
				{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
				{chg.SetHeight(6), chg.SetHeight(8), chg.SetHeight(9)},
			},
			[]change.Change{chg.SetHeight(1)},
		},
		{
			"Save in 2 batches, and load up to 9",
			9,
			[][]change.Change{
				{chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5)},
				{chg.SetHeight(6), chg.SetHeight(8), chg.SetHeight(9)},
			},
			[]change.Change{
				chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5),
				chg.SetHeight(6), chg.SetHeight(8), chg.SetHeight(9),
			},
		},
		{
			"Save in 3 batches, and load up to 8",
			8,
			[][]change.Change{
				{chg.SetHeight(1), chg.SetHeight(3)},
				{chg.SetHeight(5)},
				{chg.SetHeight(6), chg.SetHeight(8), chg.SetHeight(9)},
			},
			[]change.Change{
				chg.SetHeight(1), chg.SetHeight(3), chg.SetHeight(5),
				chg.SetHeight(6), chg.SetHeight(8),
			},
		},
	}

	for _, tt := range testcases2 {

		setup()

		for _, changes := range tt.changes {
			err := repo.Save(changes)
			assert.NoError(t, err)
		}

		changes, err := repo.LoadByNameUpToHeight(string(testNodeName1), tt.height)
		assert.NoError(t, err)
		assert.Equalf(t, tt.expected, changes, tt.name)

		cleanup()
	}
}
