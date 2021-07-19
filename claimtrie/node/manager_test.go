package node

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/node/noderepo"
	"github.com/btcsuite/btcd/claimtrie/param"
	"github.com/btcsuite/btcd/wire"

	"github.com/stretchr/testify/require"
)

var (
	out1  = NewOutPointFromString("0000000000000000000000000000000000000000000000000000000000000000:1")
	out2  = NewOutPointFromString("0000000000000000000000000000000000000000000000000000000000000000:2")
	out3  = NewOutPointFromString("0100000000000000000000000000000000000000000000000000000000000000:1")
	out4  = NewOutPointFromString("0100000000000000000000000000000000000000000000000000000000000000:2")
	name1 = []byte("name1")
	name2 = []byte("name2")
)

// verify that we can round-trip bytes to strings
func TestStringRoundTrip(t *testing.T) {

	r := require.New(t)

	data := [][]byte{
		{97, 98, 99, 0, 100, 255},
		{0xc3, 0x28},
		{0xa0, 0xa1},
		{0xe2, 0x28, 0xa1},
		{0xf0, 0x28, 0x8c, 0x28},
	}
	for _, d := range data {
		s := string(d)
		r.Equal(s, fmt.Sprintf("%s", d))
		d2 := []byte(s)
		r.Equal(len(d), len(s))
		r.Equal(d, d2)
	}
}

func TestSimpleAddClaim(t *testing.T) {

	r := require.New(t)

	param.SetNetwork(wire.TestNet)
	repo, err := noderepo.NewPebble(t.TempDir())
	r.NoError(err)

	m, err := NewBaseManager(repo)
	r.NoError(err)
	defer m.Close()

	_, err = m.IncrementHeightTo(10)
	r.NoError(err)

	chg := change.NewChange(change.AddClaim).SetName(name1).SetOutPoint(out1).SetHeight(11)
	err = m.AppendChange(chg)
	r.NoError(err)
	_, err = m.IncrementHeightTo(11)
	r.NoError(err)

	chg = chg.SetName(name2).SetOutPoint(out2).SetHeight(12)
	err = m.AppendChange(chg)
	r.NoError(err)
	_, err = m.IncrementHeightTo(12)
	r.NoError(err)

	n1, err := m.Node(name1)
	r.NoError(err)
	r.Equal(1, len(n1.Claims))
	r.NotNil(n1.Claims.find(byOut(*out1)))

	n2, err := m.Node(name2)
	r.NoError(err)
	r.Equal(1, len(n2.Claims))
	r.NotNil(n2.Claims.find(byOut(*out2)))

	err = m.DecrementHeightTo([][]byte{name2}, 11)
	r.NoError(err)
	n2, err = m.Node(name2)
	r.NoError(err)
	r.Nil(n2)

	err = m.DecrementHeightTo([][]byte{name1}, 1)
	r.NoError(err)
	n2, err = m.Node(name1)
	r.NoError(err)
	r.Nil(n2)
}

func TestSupportAmounts(t *testing.T) {

	r := require.New(t)

	param.SetNetwork(wire.TestNet)
	repo, err := noderepo.NewPebble(t.TempDir())
	r.NoError(err)

	m, err := NewBaseManager(repo)
	r.NoError(err)
	defer m.Close()

	_, err = m.IncrementHeightTo(10)
	r.NoError(err)

	chg := change.NewChange(change.AddClaim).SetName(name1).SetOutPoint(out1).SetHeight(11).SetAmount(3)
	chg.ClaimID = change.NewClaimID(*out1)
	err = m.AppendChange(chg)
	r.NoError(err)

	chg = change.NewChange(change.AddClaim).SetName(name1).SetOutPoint(out2).SetHeight(11).SetAmount(4)
	chg.ClaimID = change.NewClaimID(*out2)
	err = m.AppendChange(chg)
	r.NoError(err)

	_, err = m.IncrementHeightTo(11)
	r.NoError(err)

	chg = change.NewChange(change.AddSupport).SetName(name1).SetOutPoint(out3).SetHeight(12).SetAmount(2)
	chg.ClaimID = change.NewClaimID(*out1)
	err = m.AppendChange(chg)
	r.NoError(err)

	chg = change.NewChange(change.AddSupport).SetName(name1).SetOutPoint(out4).SetHeight(12).SetAmount(2)
	chg.ClaimID = change.NewClaimID(*out2)
	err = m.AppendChange(chg)
	r.NoError(err)

	chg = change.NewChange(change.SpendSupport).SetName(name1).SetOutPoint(out4).SetHeight(12).SetAmount(2)
	chg.ClaimID = change.NewClaimID(*out2)
	err = m.AppendChange(chg)
	r.NoError(err)

	_, err = m.IncrementHeightTo(20)
	r.NoError(err)

	n1, err := m.Node(name1)
	r.NoError(err)
	r.Equal(2, len(n1.Claims))
	r.Equal(int64(5), n1.BestClaim.Amount+n1.SupportSums[n1.BestClaim.ClaimID.Key()])
}

func TestNodeSort(t *testing.T) {

	r := require.New(t)

	param.ExtendedClaimExpirationTime = 1000

	r.True(OutPointLess(*out1, *out2))
	r.True(OutPointLess(*out1, *out3))

	n := New()
	n.Claims = append(n.Claims, &Claim{OutPoint: *out1, AcceptedAt: 3, Amount: 3, ClaimID: change.ClaimID{1}})
	n.Claims = append(n.Claims, &Claim{OutPoint: *out2, AcceptedAt: 3, Amount: 3, ClaimID: change.ClaimID{2}})
	n.handleExpiredAndActivated(3)
	n.updateTakeoverHeight(3, []byte{}, true)

	r.Equal(n.Claims.find(byOut(*out1)).OutPoint.String(), n.BestClaim.OutPoint.String())

	n.Claims = append(n.Claims, &Claim{OutPoint: *out3, AcceptedAt: 3, Amount: 3, ClaimID: change.ClaimID{3}})
	n.handleExpiredAndActivated(3)
	n.updateTakeoverHeight(3, []byte{}, true)
	r.Equal(n.Claims.find(byOut(*out1)).OutPoint.String(), n.BestClaim.OutPoint.String())
}

func TestClaimSort(t *testing.T) {

	r := require.New(t)

	param.ExtendedClaimExpirationTime = 1000

	n := New()
	n.Claims = append(n.Claims, &Claim{OutPoint: *out2, AcceptedAt: 3, Amount: 3, ClaimID: change.ClaimID{2}})
	n.Claims = append(n.Claims, &Claim{OutPoint: *out3, AcceptedAt: 3, Amount: 2, ClaimID: change.ClaimID{3}})
	n.Claims = append(n.Claims, &Claim{OutPoint: *out3, AcceptedAt: 4, Amount: 2, ClaimID: change.ClaimID{4}})
	n.Claims = append(n.Claims, &Claim{OutPoint: *out1, AcceptedAt: 3, Amount: 4, ClaimID: change.ClaimID{1}})
	n.SortClaims()

	r.Equal(int64(4), n.Claims[0].Amount)
	r.Equal(int64(3), n.Claims[1].Amount)
	r.Equal(int64(2), n.Claims[2].Amount)
	r.Equal(int32(4), n.Claims[3].AcceptedAt)
}
