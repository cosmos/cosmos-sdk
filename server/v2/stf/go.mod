module cosmossdk.io/server/v2/stf

go 1.21

replace cosmossdk.io/core => ../../../core

require (
	cosmossdk.io/core v0.11.0
	github.com/cosmos/gogoproto v1.4.12
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/btree v1.7.0
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	google.golang.org/protobuf v1.34.1
)

require (
	cosmossdk.io/log v1.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
