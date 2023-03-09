module github.com/cosmos/cosmos-sdk/orm

go 1.20

require (
	cosmossdk.io/api v0.3.1
	cosmossdk.io/core v0.6.0
	cosmossdk.io/depinject v1.0.0-alpha.3
	cosmossdk.io/errors v1.0.0-beta.7
	github.com/cosmos/cosmos-db v1.0.0-rc.1
	github.com/cosmos/cosmos-proto v1.0.0-beta.3
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.9
	github.com/iancoleman/strcase v0.2.0
	github.com/regen-network/gocuke v0.6.2
	github.com/stretchr/testify v1.8.2
	golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.29.0
	gotest.tools/v3 v3.4.0
	pgregory.net/rapid v0.5.5
)

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/alecthomas/participle/v2 v2.0.0-alpha7 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/apd/v3 v3.1.0 // indirect
	github.com/cockroachdb/errors v1.9.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v0.0.0-20230226194802-02d779ffbc46 // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/cosmos/gogoproto v1.4.6 // indirect
	github.com/cucumber/common/gherkin/go/v22 v22.0.0 // indirect
	github.com/cucumber/common/messages/go/v17 v17.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.18.0 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/linxGnu/grocksdb v1.7.15 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/onsi/gomega v1.20.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20230202175211-008b39050e57 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

// Here are the short-lived replace for orm
replace (
	cosmossdk.io/core => ../core
	cosmossdk.io/x/tx => ../x/tx
)
