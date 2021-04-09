module github.com/lbryio/chain

go 1.16

require (
	github.com/btcsuite/btcd v0.21.0-beta.0.20210526135422-418f9204f49c
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/btcsuite/go-socks v0.0.0-20170105172521-4720035b7bfd
	github.com/btcsuite/goleveldb v1.0.0
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792
	github.com/btcsuite/winsvc v1.0.0
	github.com/cockroachdb/pebble v0.0.0-20210525181856-e45797baeb78
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd/lru v1.0.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack/v5 v5.3.2
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	gorm.io/driver/postgres v1.1.0
	gorm.io/gorm v1.21.10
)

replace github.com/btcsuite/btcd => ./

replace github.com/btcsuite/btcutil => github.com/lbryio/lbcutil v1.0.3-0.20210528013734-88b0e4495162
