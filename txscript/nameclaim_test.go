package txscript

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreationParseLoopClaim(t *testing.T) {

	r := require.New(t)

	claim, err := ClaimNameScript("tester", "value")
	r.NoError(err)
	parsed, err := parseScript(claim)
	r.NoError(err)
	r.True(isClaimName(parsed))
	r.False(isSupportClaim(parsed))
	r.False(isUpdateClaim(parsed))
	script, err := DecodeClaimScript(claim)
	r.NoError(err)
	r.Equal([]byte("tester"), script.Name())
	r.Equal([]byte("value"), script.Value())
}

func TestCreationParseLoopUpdate(t *testing.T) {

	r := require.New(t)

	claimID := []byte("12345123451234512345")
	claim, err := UpdateClaimScript("tester", claimID, "value")
	r.NoError(err)
	parsed, err := parseScript(claim)
	r.NoError(err)
	r.False(isSupportClaim(parsed))
	r.False(isClaimName(parsed))
	r.True(isUpdateClaim(parsed))
	script, err := DecodeClaimScript(claim)

	r.NoError(err)
	r.Equal([]byte("tester"), script.Name())
	r.Equal(claimID, script.ClaimID())
	r.Equal([]byte("value"), script.Value())
}

func TestCreationParseLoopSupport(t *testing.T) {

	r := require.New(t)

	claimID := []byte("12345123451234512345")
	claim, err := SupportClaimScript("tester", claimID, []byte("value"))
	r.NoError(err)
	parsed, err := parseScript(claim)
	r.NoError(err)
	r.True(isSupportClaim(parsed))
	r.False(isClaimName(parsed))
	r.False(isUpdateClaim(parsed))
	script, err := DecodeClaimScript(claim)

	r.NoError(err)
	r.Equal([]byte("tester"), script.Name())
	r.Equal(claimID, script.ClaimID())
	r.Equal([]byte("value"), script.Value())

	claim, err = SupportClaimScript("tester", claimID, nil)
	r.NoError(err)
	script, err = DecodeClaimScript(claim)
	r.NoError(err)

	r.Equal([]byte("tester"), script.Name())
	r.Equal(claimID, script.ClaimID())
	r.Nil(script.Value())
}
