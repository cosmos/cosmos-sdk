module github.com/cosmos/cosmos-sdk

go 1.13

require (
	github.com/bartekn/go-bip39 v0.0.0-20171116152956-a05967ea095d
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.23.4 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/ledger-cosmos-go v0.13.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/mattn/go-isatty v0.0.17
	github.com/pelletier/go-toml v1.9.5
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.13.0
	github.com/stretchr/testify v1.8.3
	github.com/stumble/gorocksdb v0.0.3 // indirect
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/iavl v0.12.4
	github.com/tendermint/tendermint v0.34.21
	github.com/tendermint/tm-db v0.6.7
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/tendermint/tendermint => github.com/maticnetwork/tendermint v0.26.0-dev0.0.20231004080423-ab02946e7b7d

replace github.com/tendermint/tm-db => github.com/tendermint/tm-db v0.2.0
