go 1.23.1

module github.com/cosmos/cosmos-sdk

require (
	cosmossdk.io/api v0.7.5
	cosmossdk.io/collections v0.4.0
	cosmossdk.io/core v1.0.0-alpha.3
	cosmossdk.io/core/testing v0.0.0-20240923163230-04da382a9f29
	cosmossdk.io/depinject v1.0.0
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/log v1.4.1
	cosmossdk.io/math v1.3.0
	cosmossdk.io/schema v0.3.0
	cosmossdk.io/store v1.1.1-0.20240418092142-896cdf1971bc
	cosmossdk.io/x/bank v0.0.0-20240226161501-23359a0b6d91
	cosmossdk.io/x/staking v0.0.0-00010101000000-000000000000
	cosmossdk.io/x/tx v0.13.3
	github.com/99designs/keyring v1.2.2
	github.com/bgentry/speakeasy v0.2.0
	github.com/cometbft/cometbft v1.0.0-rc1.0.20240908111210-ab0be101882f
	github.com/cometbft/cometbft/api v1.0.0-rc.1
	github.com/cosmos/btcutil v1.0.5
	github.com/cosmos/cosmos-db v1.0.3-0.20240911104526-ddc3f09bfc22
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/crypto v0.1.2
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/iavl v0.17.2
	github.com/cosmos/ledger-cosmos-go v0.11.1
	github.com/enigmampc/btcutil v1.0.3-0.20200723161021-e2fb6adb2a25
	github.com/gogo/gateway v1.1.0
	github.com/gogo/protobuf v1.3.3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hdevalence/ed25519consensus v0.0.0-20210204194344-59a8610d2b87
	github.com/improbable-eng/grpc-web v0.14.1
	github.com/jhump/protoreflect v1.9.0
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.2 // indirect
	github.com/magiconair/properties v1.8.5
	github.com/mattn/go-isatty v0.0.14
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.13.0 // indirect
	github.com/otiai10/copy v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.29.0
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.14
	github.com/tendermint/tm-db v0.6.4
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

// latest grpc doesn't work with with our modified proto compiler, so we need to enforce
// the following version across all dependencies.
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

// The following is to test committingClient (allowing parallel queries during
// write transactions):
replace github.com/tendermint/tendermint => github.com/agoric-labs/tendermint v0.34.14-alpha.agoric.1

replace github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76

// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0
