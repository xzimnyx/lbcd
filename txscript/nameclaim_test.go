package txscript

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreationParseLoopClaim(t *testing.T) {

	r := require.New(t)

	claim, err := ClaimNameScript("tester", "value")
	r.NoError(err)
	parsed, closer, err := parseScript(claim)
	defer closer()
	r.NoError(err)
	r.True(isClaimName(parsed))
	r.False(isSupportClaim(parsed))
	r.False(isUpdateClaim(parsed))
	script, closer2, err := DecodeClaimScript(claim)
	defer closer2()
	r.NoError(err)
	r.Equal([]byte("tester"), script.Name())
	r.Equal([]byte("value"), script.Value())
}

func TestCreationParseLoopUpdate(t *testing.T) {

	r := require.New(t)

	claimID := []byte("12345123451234512345")
	claim, err := UpdateClaimScript("tester", claimID, "value")
	r.NoError(err)
	parsed, closer, err := parseScript(claim)
	defer closer()
	r.NoError(err)
	r.False(isSupportClaim(parsed))
	r.False(isClaimName(parsed))
	r.True(isUpdateClaim(parsed))
	script, closer2, err := DecodeClaimScript(claim)
	defer closer2()

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
	parsed, closer, err := parseScript(claim)
	defer closer()

	r.NoError(err)
	r.True(isSupportClaim(parsed))
	r.False(isClaimName(parsed))
	r.False(isUpdateClaim(parsed))
	script, closer2, err := DecodeClaimScript(claim)
	defer closer2()

	r.NoError(err)
	r.Equal([]byte("tester"), script.Name())
	r.Equal(claimID, script.ClaimID())
	r.Equal([]byte("value"), script.Value())

	claim, err = SupportClaimScript("tester", claimID, nil)
	r.NoError(err)
	script, closer, err = DecodeClaimScript(claim)
	defer closer()
	r.NoError(err)

	r.Equal([]byte("tester"), script.Name())
	r.Equal(claimID, script.ClaimID())
	r.Nil(script.Value())
}

func TestInvalidChars(t *testing.T) {
	r := require.New(t)

	script, err := ClaimNameScript("tester", "value")
	r.NoError(err)
	r.NoError(AllClaimsAreSane(script, true))

	for i := range []byte(illegalChars) {
		script, err := ClaimNameScript("a"+illegalChars[i:i+1], "value")
		r.NoError(err)
		r.Error(AllClaimsAreSane(script, true))
	}
}
