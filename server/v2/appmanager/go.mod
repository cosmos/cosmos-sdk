module cosmossdk.io/server/v2/appmanager

go 1.21

replace (
	cosmossdk.io/server/v2/core => ../core
	cosmossdk.io/server/v2/stf => ../stf
	cosmossdk.io/store/v2 => ../../../store
)

require (
	cosmossdk.io/core v0.12.0
	cosmossdk.io/log v1.2.1
	cosmossdk.io/server/v2/core v0.0.0-00010101000000-000000000000
	cosmossdk.io/server/v2/stf v0.0.0-00010101000000-000000000000
	cosmossdk.io/store/v2 v2.0.0-00010101000000-000000000000
	github.com/cosmos/cosmos-db v1.0.0
	github.com/stretchr/testify v1.8.4
	google.golang.org/protobuf v1.32.0
)

require (
	cosmossdk.io/errors v1.0.0 // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v0.0.0-20231102162011-844f0582c2eb // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cometbft/cometbft v0.38.2 // indirect
	github.com/cometbft/cometbft/api v1.0.0-alpha.1 // indirect
	github.com/cosmos/gogoproto v1.4.11 // indirect
	github.com/cosmos/iavl v1.0.0 // indirect
	github.com/cosmos/ics23/go v0.10.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/emicklei/dot v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.25.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-metrics v0.5.3 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.8.10 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20220708102147-0a8a51822cae // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/rs/zerolog v1.31.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
	google.golang.org/grpc v1.60.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
