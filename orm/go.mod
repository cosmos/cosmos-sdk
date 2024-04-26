module cosmossdk.io/orm

go 1.21

require (
	cosmossdk.io/api v0.7.4
	cosmossdk.io/core v0.12.1-0.20231114100755-569e3ff6a0d7
	cosmossdk.io/depinject v1.0.0-alpha.4
	cosmossdk.io/errors v1.0.1
	github.com/cosmos/cosmos-db v1.0.2
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.6.0
	github.com/iancoleman/strcase v0.3.0
	github.com/regen-network/gocuke v1.1.1
	github.com/stretchr/testify v1.9.0
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225
	google.golang.org/grpc v1.63.2
	google.golang.org/protobuf v1.33.0
	gotest.tools/v3 v3.5.1
	pgregory.net/rapid v1.1.0
)

require (
	cosmossdk.io/log v1.3.1 // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.0 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cosmos/gogoproto v1.4.12 // indirect
	github.com/cucumber/gherkin/go/v27 v27.0.0 // indirect
	github.com/cucumber/messages/go/v22 v22.0.0 // indirect
	github.com/cucumber/tag-expressions/go/v6 v6.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.8.14 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/onsi/gomega v1.20.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240401170217-c3f982113cda // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace (
	cosmossdk.io/core => ../core
	cosmossdk.io/depinject => ../depinject
)
