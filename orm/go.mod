module cosmossdk.io/orm

go 1.23

require (
	cosmossdk.io/api v0.7.5
	cosmossdk.io/core v1.0.0-alpha.3
	cosmossdk.io/core/testing v0.0.0-00010101000000-000000000000
	cosmossdk.io/depinject v1.0.0
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/store v1.1.1
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.6.0
	github.com/iancoleman/strcase v0.3.0
	github.com/regen-network/gocuke v1.1.1
	github.com/stretchr/testify v1.9.0
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842
	google.golang.org/grpc v1.66.2
	google.golang.org/protobuf v1.34.2
	gotest.tools/v3 v3.5.1
	pgregory.net/rapid v1.1.0
)

require (
	cosmossdk.io/schema v0.3.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
	github.com/cosmos/gogoproto v1.7.0 // indirect
	github.com/cucumber/gherkin/go/v27 v27.0.0 // indirect
	github.com/cucumber/messages/go/v22 v22.0.0 // indirect
	github.com/cucumber/tag-expressions/go/v6 v6.1.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/onsi/gomega v1.20.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240604185151-ef581f913117 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace cosmossdk.io/core/testing => ../core/testing

replace cosmossdk.io/store => ../store
