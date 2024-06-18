module cosmossdk.io/runtime/v2

go 1.22.2

// server v2 integration
replace (
	cosmossdk.io/api => ../../api
	cosmossdk.io/core => ../../core
	cosmossdk.io/depinject => ../../depinject
	cosmossdk.io/log => ../../log
	cosmossdk.io/server/v2/appmanager => ../../server/v2/appmanager
	cosmossdk.io/server/v2/stf => ../../server/v2/stf
	cosmossdk.io/store/v2 => ../../store/v2
	cosmossdk.io/x/accounts => ../../x/accounts
	cosmossdk.io/x/auth => ../../x/auth
	cosmossdk.io/x/bank => ../../x/bank
	cosmossdk.io/x/consensus => ../../x/consensus
	cosmossdk.io/x/distribution => ../../x/distribution
	cosmossdk.io/x/staking => ../../x/staking
	cosmossdk.io/x/tx => ../../x/tx
	github.com/cosmos/cosmos-sdk => ../..
)

require (
	cosmossdk.io/api v0.7.5
	cosmossdk.io/core v0.12.1-0.20231114100755-569e3ff6a0d7
	cosmossdk.io/depinject v1.0.0-alpha.4
	cosmossdk.io/server/v2/appmanager v0.0.0-00010101000000-000000000000
	cosmossdk.io/server/v2/stf v0.0.0-00010101000000-000000000000
	cosmossdk.io/store/v2 v2.0.0-00010101000000-000000000000
	cosmossdk.io/x/tx v0.13.3
	github.com/cosmos/gogoproto v1.5.0
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2
)

require (
	buf.build/gen/go/cometbft/cometbft/protocolbuffers/go v1.34.1-20240312114316-c0d3497e35d6.1 // indirect
	buf.build/gen/go/cosmos/gogo-proto/protocolbuffers/go v1.34.1-20240130113600-88ef6483f90f.1 // indirect
	cosmossdk.io/errors v1.0.1 // indirect
	cosmossdk.io/log v1.3.1 // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.0 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cosmos/cosmos-db v1.0.2 // indirect
	github.com/cosmos/cosmos-proto v1.0.0-beta.5 // indirect
	github.com/cosmos/iavl v1.2.0 // indirect
	github.com/cosmos/ics23/go v0.10.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/dot v1.6.2 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.3 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.8.14 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/onsi/gomega v1.28.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.54.0 // indirect
	github.com/prometheus/procfs v0.14.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240513163218-0867130af1f8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
