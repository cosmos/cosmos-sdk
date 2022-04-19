go 1.16

module github.com/cosmos/cosmos-sdk

require (
	github.com/99designs/keyring v1.2.1
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/armon/go-metrics v0.3.10
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/confio/ics23/go v0.6.6
	github.com/cosmos/btcutil v1.0.4
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/iavl v0.17.3
	github.com/cosmos/ledger-cosmos-go v0.11.1
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/gogo/gateway v1.1.0
	github.com/gogo/protobuf v1.3.3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/gofuzz v1.1.1-0.20200604201612-c04b05f3adfa // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/magiconair/properties v1.8.6
	github.com/mattn/go-isatty v0.0.14
	github.com/otiai10/copy v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/common v0.32.1
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rs/zerolog v1.21.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.11.0
	github.com/stretchr/testify v1.7.1
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.19
	github.com/tendermint/tm-db v0.6.7
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
	google.golang.org/genproto v0.0.0-20220414192740-2d67ff6cf2b4
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v2 v2.4.0

)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
