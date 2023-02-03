module cosmossdk.io/store

go 1.19

require (
	cosmossdk.io/errors v1.0.0-beta.7
	github.com/armon/go-metrics v0.4.1
	github.com/confio/ics23/go v0.9.0
	github.com/cosmos/cosmos-db v0.0.0-20230119180254-161cf3632b7c
	github.com/cosmos/gogoproto v1.4.4
	github.com/cosmos/iavl v0.20.0-alpha3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/spf13/cast v1.5.0
	github.com/stretchr/testify v1.8.1
	github.com/tendermint/tendermint v0.37.0-rc2
	github.com/tidwall/btree v1.6.0
	golang.org/x/exp v0.0.0-20230118134722-a68e582fa157
	google.golang.org/genproto v0.0.0-20230125152338-dcaf20b6aeaa // indirect
	google.golang.org/grpc v1.52.3 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gotest.tools/v3 v3.4.0
)

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cockroachdb/errors v1.9.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v0.0.0-20220817183557-09c6e030a677 // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.17.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.7.10 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.8.1
