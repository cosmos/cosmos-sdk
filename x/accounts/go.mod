module cosmossdk.io/x/accounts

go 1.21

require (
	cosmossdk.io/api v0.7.1
	cosmossdk.io/collections v0.4.0
	cosmossdk.io/core v0.11.0
	github.com/cosmos/gogoproto v1.4.11
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.58.2
	google.golang.org/protobuf v1.31.0
)

require (
	cosmossdk.io/depinject v1.0.0-alpha.4 // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v0.0.0-20230525220056-bb4fc9527b3b // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cosmos/cosmos-db v1.0.0 // indirect
	github.com/cosmos/cosmos-proto v1.0.0-beta.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/getsentry/sentry-go v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.7.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto v0.0.0-20230913181813-007df8e322eb // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230803162519-f966b187b2e5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230920204549-e6e6cdab5c13 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// REMOVE post v1 release.
replace cosmossdk.io/api => ../../api
