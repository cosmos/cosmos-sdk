module cosmossdk.io/server/v2

go 1.21

replace (
	cosmossdk.io/api => ../../api
	cosmossdk.io/core => ../../core
	cosmossdk.io/core/testing => ../../core/testing
	cosmossdk.io/server/v2/appmanager => ./appmanager
	cosmossdk.io/server/v2/stf => ./stf
	cosmossdk.io/x/tx => ../../x/tx
)

require (
	cosmossdk.io/api v0.7.5
	cosmossdk.io/core v0.12.1-0.20231114100755-569e3ff6a0d7
	cosmossdk.io/core/testing v0.0.0-00010101000000-000000000000
	cosmossdk.io/log v1.3.1
	cosmossdk.io/server/v2/appmanager v0.0.0-00010101000000-000000000000
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/gogogateway v1.2.0
	github.com/cosmos/gogoproto v1.5.0
	github.com/golang/protobuf v1.5.4
	github.com/gorilla/mux v1.8.1
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-hclog v1.6.2
	github.com/hashicorp/go-metrics v0.5.3
	github.com/hashicorp/go-plugin v1.6.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pelletier/go-toml/v2 v2.2.2
	github.com/prometheus/client_golang v1.19.1
	github.com/prometheus/common v0.55.0
	github.com/rs/zerolog v1.33.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.9.0
	golang.org/x/exp v0.0.0-20240531132922-fd00a4e0eefc
	golang.org/x/sync v0.7.0
	google.golang.org/grpc v1.64.1
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/DataDog/datadog-go v3.2.0+incompatible // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jhump/protoreflect v1.15.3 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240709173604-40e1e62336c5 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
