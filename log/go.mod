module cosmossdk.io/log

go 1.20

require (
	github.com/rs/zerolog v1.29.0
	gotest.tools/v3 v3.4.0
)

require (
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	golang.org/x/sys v0.5.0 // indirect
)

// TODO wait for https://github.com/rs/zerolog/pull/527 to be merged
replace github.com/rs/zerolog => github.com/julienrbrt/zerolog v0.0.0-20230227160104-8fd4f8ce93ff
