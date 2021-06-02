package param

const (
	DefaultMaxActiveDelay    int32 = 4032
	DefaultActiveDelayFactor int32 = 32
)

// https://lbry.io/news/hf1807
const (
	DefaultOriginalClaimExpirationTime       int32 = 262974
	DefaultExtendedClaimExpirationTime       int32 = 2102400
	DefaultExtendedClaimExpirationForkHeight int32 = 400155
)

var (
	MaxActiveDelay                    = DefaultMaxActiveDelay
	ActiveDelayFactor                 = DefaultActiveDelayFactor
	OriginalClaimExpirationTime       = DefaultOriginalClaimExpirationTime
	ExtendedClaimExpirationTime       = DefaultExtendedClaimExpirationTime
	ExtendedClaimExpirationForkHeight = DefaultExtendedClaimExpirationForkHeight
)

// https://lbry.com/news/hf1903

// https://lbry.com/news/hf1910
