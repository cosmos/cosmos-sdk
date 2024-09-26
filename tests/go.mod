module github.com/cosmos/cosmos-sdk/tests

go 1.23.1

require (
	cosmossdk.io/api v0.7.6
	cosmossdk.io/collections v0.4.0
	cosmossdk.io/core v1.0.0-alpha.3
	cosmossdk.io/depinject v1.0.0
	cosmossdk.io/log v1.4.1
	cosmossdk.io/math v1.3.0
	cosmossdk.io/simapp v0.0.0-20230309163709-87da587416ba
	cosmossdk.io/store v1.1.1
	cosmossdk.io/x/evidence v0.0.0-20230613133644-0a778132a60f
	cosmossdk.io/x/feegrant v0.0.0-20230613133644-0a778132a60f
	cosmossdk.io/x/nft v0.0.0-20230613133644-0a778132a60f // indirect
	cosmossdk.io/x/protocolpool v0.0.0-20230925135524-a1bc045b3190
	cosmossdk.io/x/tx v0.13.5
	cosmossdk.io/x/upgrade v0.0.0-20230613133644-0a778132a60f
	github.com/cometbft/cometbft v1.0.0-rc1.0.20240908111210-ab0be101882f
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	// this version is not used as it is always replaced by the latest Cosmos SDK version
	github.com/cosmos/cosmos-sdk v0.53.0
	github.com/cosmos/gogoproto v1.7.0
	github.com/golang/mock v1.6.0
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.67.0
	google.golang.org/protobuf v1.34.2
	gotest.tools/v3 v3.5.1
	pgregory.net/rapid v1.1.0
)

require (
	cosmossdk.io/core/testing v0.0.0-20240923163230-04da382a9f29
	cosmossdk.io/x/accounts v0.0.0-20240913065641-0064ccbce64e
	cosmossdk.io/x/accounts/defaults/base v0.0.0-00010101000000-000000000000
	cosmossdk.io/x/accounts/defaults/lockup v0.0.0-20240417181816-5e7aae0db1f5
	cosmossdk.io/x/accounts/defaults/multisig v0.0.0-00010101000000-000000000000
	cosmossdk.io/x/authz v0.0.0-00010101000000-000000000000
	cosmossdk.io/x/bank v0.0.0-20240226161501-23359a0b6d91
	cosmossdk.io/x/consensus v0.0.0-00010101000000-000000000000
	cosmossdk.io/x/distribution v0.0.0-20240227221813-a248d05f70f4
	cosmossdk.io/x/gov v0.0.0-20231113122742-912390d5fc4a
	cosmossdk.io/x/group v0.0.0-00010101000000-000000000000
	cosmossdk.io/x/mint v0.0.0-00010101000000-000000000000
	cosmossdk.io/x/slashing v0.0.0-00010101000000-000000000000
	cosmossdk.io/x/staking v0.0.0-20240226161501-23359a0b6d91
	github.com/cometbft/cometbft/api v1.0.0-rc.1
	github.com/google/go-cmp v0.6.0
	github.com/google/gofuzz v1.2.0
	github.com/jhump/protoreflect v1.17.0
	github.com/rs/zerolog v1.33.0
	github.com/spf13/viper v1.19.0
)

require (
	buf.build/gen/go/cometbft/cometbft/protocolbuffers/go v1.34.2-20240701160653-fedbb9acfd2f.2 // indirect
	buf.build/gen/go/cosmos/gogo-proto/protocolbuffers/go v1.34.2-20240130113600-88ef6483f90f.2 // indirect
	cloud.google.com/go v0.115.1 // indirect
	cloud.google.com/go/auth v0.8.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.4 // indirect
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	cloud.google.com/go/iam v1.1.13 // indirect
	cloud.google.com/go/storage v1.43.0 // indirect
	cosmossdk.io/client/v2 v2.0.0-20230630094428-02b760776860 // indirect
	cosmossdk.io/errors v1.0.1 // indirect
	cosmossdk.io/schema v0.3.0 // indirect
	cosmossdk.io/x/circuit v0.0.0-20230613133644-0a778132a60f // indirect
	cosmossdk.io/x/epochs v0.0.0-20240522060652-a1ae4c3e0337 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/DataDog/datadog-go v4.8.3+incompatible // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/bgentry/speakeasy v0.2.0 // indirect
	github.com/bits-and-blooms/bitset v1.10.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cockroachdb/apd/v2 v2.0.2 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240606204812-0bbfbd93a7ce // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.2 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cometbft/cometbft-db v0.15.0 // indirect
	github.com/cosmos/btcutil v1.0.5 // indirect
	github.com/cosmos/cosmos-db v1.0.3-0.20240911104526-ddc3f09bfc22 // indirect
	github.com/cosmos/crypto v0.1.2 // indirect
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/cosmos/gogogateway v1.2.0 // indirect
	github.com/cosmos/iavl v1.3.0 // indirect
	github.com/cosmos/ics23/go v0.11.0 // indirect
	github.com/cosmos/ledger-cosmos-go v0.13.3 // indirect
	github.com/danieljoos/wincred v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/dgraph-io/badger/v4 v4.3.0 // indirect
	github.com/dgraph-io/ristretto v0.1.2-0.20240116140435-c67e07994f91 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/emicklei/dot v1.6.2 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/google/orderedcode v0.0.1 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter v1.7.6 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.3 // indirect
	github.com/hashicorp/go-plugin v1.6.1 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/hdevalence/ed25519consensus v0.2.0 // indirect
	github.com/huandu/skiplist v1.2.1 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/linxGnu/grocksdb v1.9.3 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/manifoldco/promptui v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20230904125328-1f23a7beb09a // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.20.4 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.59.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/supranational/blst v0.3.13 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/zondax/hid v0.9.2 // indirect
	github.com/zondax/ledger-go v0.14.3 // indirect
	gitlab.com/yawning/secp256k1-voi v0.0.0-20230925100816-f2616030848b // indirect
	gitlab.com/yawning/tuplehash v0.0.0-20230713102510-df83abbf9a02 // indirect
	go.etcd.io/bbolt v1.4.0-alpha.0.0.20240404170359-43604f3112c5 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.53.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/exp v0.0.0-20240531132922-fd00a4e0eefc // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/term v0.24.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/api v0.192.0 // indirect
	google.golang.org/genproto v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240924160255-9d4c2d233b61 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

// Here are the short-lived replace from the SimApp
// Replace here are pending PRs, or version to be tagged
// replace (
// 	<temporary replace>
// )

// SimApp on main always tests the latest extracted SDK modules importing the sdk
replace (
	cosmossdk.io/api => ../api
	cosmossdk.io/client/v2 => ../client/v2
	cosmossdk.io/collections => ../collections
	cosmossdk.io/store => ../store
	cosmossdk.io/x/accounts => ../x/accounts
	cosmossdk.io/x/accounts/defaults/base => ../x/accounts/defaults/base
	cosmossdk.io/x/accounts/defaults/lockup => ../x/accounts/defaults/lockup
	cosmossdk.io/x/accounts/defaults/multisig => ../x/accounts/defaults/multisig
	cosmossdk.io/x/authz => ../x/authz
	cosmossdk.io/x/bank => ../x/bank
	cosmossdk.io/x/circuit => ../x/circuit
	cosmossdk.io/x/consensus => ../x/consensus
	cosmossdk.io/x/distribution => ../x/distribution
	cosmossdk.io/x/epochs => ../x/epochs
	cosmossdk.io/x/evidence => ../x/evidence
	cosmossdk.io/x/feegrant => ../x/feegrant
	cosmossdk.io/x/gov => ../x/gov
	cosmossdk.io/x/group => ../x/group
	cosmossdk.io/x/mint => ../x/mint
	cosmossdk.io/x/nft => ../x/nft
	cosmossdk.io/x/params => ../x/params
	cosmossdk.io/x/protocolpool => ../x/protocolpool
	cosmossdk.io/x/slashing => ../x/slashing
	cosmossdk.io/x/staking => ../x/staking
	cosmossdk.io/x/tx => ../x/tx
	cosmossdk.io/x/upgrade => ../x/upgrade
)

// Below are the long-lived replace for tests.
replace (
	// We always want to test against the latest version of the simapp.
	cosmossdk.io/simapp => ../simapp
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	// We always want to test against the latest version of the SDK.
	github.com/cosmos/cosmos-sdk => ../.
)
