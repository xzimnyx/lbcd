package param

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"path/filepath"
)

var (
	MaxActiveDelay    int32
	ActiveDelayFactor int32

	MaxNodeManagerCacheSize int

	OriginalClaimExpirationTime       int32
	ExtendedClaimExpirationTime       int32
	ExtendedClaimExpirationForkHeight int32

	MaxRemovalWorkaroundHeight int32

	NormalizedNameForkHeight    int32
	AllClaimsInMerkleForkHeight int32

	ClaimtrieDataFolder string
)

func SetNetwork(net wire.BitcoinNet, subfolder string) {
	MaxActiveDelay = 4032
	ActiveDelayFactor = 32
	MaxNodeManagerCacheSize = 16000
	appDir := btcutil.AppDataDir("chain", false)
	ClaimtrieDataFolder = filepath.Join(appDir, "data", subfolder, "claim_dbs")

	switch net {
	case wire.MainNet:
		OriginalClaimExpirationTime = 262974
		ExtendedClaimExpirationTime = 2102400
		ExtendedClaimExpirationForkHeight = 400155 // https://lbry.io/news/hf1807
		MaxRemovalWorkaroundHeight = 658300
		NormalizedNameForkHeight = 539940    // targeting 21 March 2019}, https://lbry.com/news/hf1903
		AllClaimsInMerkleForkHeight = 658309 // targeting 30 Oct 2019}, https://lbry.com/news/hf1910
	case wire.TestNet3:
		OriginalClaimExpirationTime = 262974
		ExtendedClaimExpirationTime = 2102400
		ExtendedClaimExpirationForkHeight = 1
		MaxRemovalWorkaroundHeight = 100
		NormalizedNameForkHeight = 1
		AllClaimsInMerkleForkHeight = 109
	case wire.TestNet, wire.SimNet: // "regtest"
		OriginalClaimExpirationTime = 500
		ExtendedClaimExpirationTime = 600
		ExtendedClaimExpirationForkHeight = 800
		MaxRemovalWorkaroundHeight = -1
		NormalizedNameForkHeight = 250
		AllClaimsInMerkleForkHeight = 349
	}
}
