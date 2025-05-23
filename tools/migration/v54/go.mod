module github.com/cosmos/cosmos-sdk/tools/migration/v54

go 1.23.2

replace github.com/cosmos/cosmos-sdk/tools/migrate => ../

require (
	github.com/cosmos/cosmos-sdk/tools/migrate v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.34.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
)
