module cosmossdk.io/x/circuit

go 1.21.0

require (
	cosmossdk.io/collections v0.4.0
	cosmossdk.io/core v0.12.1-0.20231114100755-569e3ff6a0d7
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/server/v2/stf v0.0.0-00010101000000-000000000000
	github.com/cosmos/gogoproto v1.4.12
	github.com/gogo/protobuf v1.3.2
	google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de
)

require (
	cosmossdk.io/log v1.3.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/sys v0.19.0 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace cosmossdk.io/core => ../../core

replace cosmossdk.io/collections => ../../collections

replace cosmossdk.io/server/v2/stf => ../../server/v2/stf

replace cosmossdk.io/errors => ../../errors
