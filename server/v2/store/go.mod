module cosmossdk.io/server/v2/store

go 1.22.2

replace (
	cosmossdk.io/api => ../../../api
	cosmossdk.io/core => ../../../core
	cosmossdk.io/core/testing => ../../../core/testing
	cosmossdk.io/depinject => ../../../depinject
	cosmossdk.io/log => ../../../log
	cosmossdk.io/server/v2 => ../
	cosmossdk.io/server/v2/appmanager => ../appmanager
	cosmossdk.io/store => ../../../store
	cosmossdk.io/store/v2 => ../../../store/v2
	cosmossdk.io/x/accounts => ../../../x/accounts
	cosmossdk.io/x/auth => ../../../x/auth
	cosmossdk.io/x/bank => ../../../x/bank
	cosmossdk.io/x/consensus => ../../../x/consensus
	cosmossdk.io/x/staking => ../../../x/staking
	github.com/cosmos/cosmos-sdk => ../../../
)

require (
	cosmossdk.io/core v0.12.1-0.20231114100755-569e3ff6a0d7
	cosmossdk.io/log v1.3.1
	cosmossdk.io/server/v2 v2.0.0-00010101000000-000000000000
	cosmossdk.io/store v1.1.1-0.20240418092142-896cdf1971bc
	cosmossdk.io/store/v2 v2.0.0-00010101000000-000000000000
	github.com/cosmos/cosmos-sdk v0.51.0
	github.com/spf13/cast v1.6.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
)

require (
	cosmossdk.io/errors v1.0.1 // indirect
	cosmossdk.io/server/v2/appmanager v0.0.0-00010101000000-000000000000 // indirect
	github.com/cosmos/gogoproto v1.5.0 // indirect
	github.com/cosmos/ics23/go v0.10.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.3 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/exp v0.0.0-20240531132922-fd00a4e0eefc // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240515191416-fc5f0ca64291 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
