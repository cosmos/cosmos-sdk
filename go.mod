go 1.15

module github.com/cosmos/cosmos-sdk

require (
	github.com/99designs/keyring v1.1.6
	github.com/armon/go-metrics v0.3.10
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/coinbase/rosetta-sdk-go v0.7.0
	github.com/confio/ics23/go v0.6.6
	github.com/cosmos/btcutil v1.0.4
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/iavl v0.17.3
	github.com/cosmos/ledger-cosmos-go v0.11.1
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
	github.com/magiconair/properties v1.8.5
	github.com/mattn/go-isatty v0.0.14
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/otiai10/copy v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/common v0.32.1
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.16
	github.com/tendermint/tm-db v0.6.6
	golang.org/x/crypto v0.0.0-20210915214749-c084706c2272
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

// latest grpc doesn't work with with our modified proto compiler, so we need to enforce
// the following version across all dependencies.
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76

// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0
