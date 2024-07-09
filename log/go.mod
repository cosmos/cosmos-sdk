module cosmossdk.io/log

go 1.20

require (
	cosmossdk.io/core v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.33.0
	gotest.tools/v3 v3.5.1
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.22.0 // indirect
)

replace cosmossdk.io/core => ../core
