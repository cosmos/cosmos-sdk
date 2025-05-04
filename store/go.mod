module cosmossdk.io/store

go 1.23.0

toolchain go1.23.8

require (
	cosmossdk.io/errors v1.0.0
	cosmossdk.io/log v1.4.1
	cosmossdk.io/math v1.4.0
	github.com/cometbft/cometbft v0.38.12
	github.com/cosmos/cosmos-db v1.0.2
	github.com/cosmos/gogoproto v1.7.0
	github.com/cosmos/iavl v1.2.0
	github.com/cosmos/ics23/go v0.11.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-metrics v0.5.1
	github.com/hashicorp/go-plugin v1.5.2
	github.com/hashicorp/golang-lru v1.0.2
	github.com/spf13/cast v1.6.0 // indirect
	github.com/stretchr/testify v1.10.0
	github.com/tidwall/btree v1.7.0
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.5
	gotest.tools/v3 v3.5.1
)

require github.com/cometbft/cometbft/api v1.0.0

require (
	github.com/DataDog/zstd v1.5.6 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240816210425-c5d0cb0b6fc0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20241215232642-bb51bb14a506 // indirect
	github.com/cockroachdb/pebble v1.1.4 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/dot v1.6.1 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/getsentry/sentry-go v0.31.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/jhump/protoreflect v1.15.3 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.9.8 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.21.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	// Use CometBFT v1.0.1 with Mempool lanes and DOG
	github.com/cometbft/cometbft => github.com/InjectiveLabs/cometbft v1.0.1-inj
	github.com/cometbft/cometbft/api => github.com/injectivelabs/cometbft/api v1.0.1-0.20250315062455-e9e4c8a0ecb9
)
