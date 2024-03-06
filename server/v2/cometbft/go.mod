module cosmossdk.io/server/v2/cometbft

go 1.21

replace (
	cosmossdk.io/core => ../../../core
	cosmossdk.io/server/v2 => ../
	cosmossdk.io/server/v2/appmanager => ../appmanager
	cosmossdk.io/server/v2/core => ../core
	cosmossdk.io/server/v2/stf => ../stf
	cosmossdk.io/store/v2 => ../../../store
	github.com/cometbft/cometbft/api => github.com/cometbft/cometbft/api v1.0.0-alpha.1
	github.com/cosmos/iavl => github.com/cosmos/iavl v1.0.0-beta.1.0.20240125174944-11ba4961dae9
)

require (
	buf.build/gen/go/tendermint/tendermint/protocolbuffers/go v1.32.0-20231117195010-33ed361a9051.1
	cosmossdk.io/api v0.7.3
	cosmossdk.io/core v0.12.0
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/log v1.3.1
	cosmossdk.io/server/v2 v2.0.0-00010101000000-000000000000
	cosmossdk.io/server/v2/appmanager v0.0.0-00010101000000-000000000000
	cosmossdk.io/server/v2/core v0.0.0-00010101000000-000000000000
	cosmossdk.io/store/v2 v2.0.0-00010101000000-000000000000
	github.com/cometbft/cometbft v0.38.5
	github.com/cosmos/gogoproto v1.4.11
	github.com/cosmos/ics23/go v0.10.0
	google.golang.org/protobuf v1.32.0
)

require (
	buf.build/gen/go/cosmos/gogo-proto/protocolbuffers/go v1.32.0-20230509103710-5e5b9fdd0180.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cometbft/cometbft-db v0.9.1 // indirect
	github.com/cosmos/cosmos-proto v1.0.0-beta.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/golang/glog v1.2.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/go-hclog v1.6.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.3 // indirect
	github.com/hashicorp/go-plugin v1.6.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/linxGnu/grocksdb v1.8.12 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20230904125328-1f23a7beb09a // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/petermattis/goid v0.0.0-20230904192822-1876fd5063bc // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	go.etcd.io/bbolt v1.3.8 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240205150955-31a09d347014 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240221002015-b0ce06bbee7c // indirect
	google.golang.org/grpc v1.62.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
