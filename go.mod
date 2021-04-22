go 1.13

module github.com/cosmos/cosmos-sdk

require (
	github.com/99designs/keyring v1.1.6
	github.com/bartekn/go-bip39 v0.0.0-20171116152956-a05967ea095d
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/cosmos/ledger-cosmos-go v0.11.1
	github.com/ethereum/go-ethereum v1.9.25
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.3.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/mattn/go-isatty v0.0.12
	github.com/pelletier/go-toml v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.6
	github.com/spf13/afero v1.2.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/iavl v0.14.1
	github.com/tendermint/tendermint v0.33.9
	github.com/tendermint/tm-db v0.5.1
	golang.org/x/net v0.0.0-20201010224723-4f7140c49acb // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
