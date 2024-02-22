module cosmossdk.io/server/v2

go 1.21

replace (
	cosmossdk.io/core => ../../core
	cosmossdk.io/server/v2/core => ./core
	cosmossdk.io/server/v2/stf => ./stf
)

require (
	cosmossdk.io/api v0.7.2
	cosmossdk.io/core v0.11.0
	cosmossdk.io/log v1.3.1
	cosmossdk.io/server/v2/core v0.0.0-00010101000000-000000000000
	github.com/cosmos/cosmos-proto v1.0.0-beta.3
	github.com/cosmos/gogogateway v1.2.0
	github.com/cosmos/gogoproto v1.4.11
	github.com/golang/protobuf v1.5.3
	github.com/gorilla/mux v1.8.1
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-hclog v1.6.2
	github.com/hashicorp/go-metrics v0.5.1
	github.com/hashicorp/go-plugin v1.6.0
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/common v0.45.0
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.61.1
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/DataDog/datadog-go v3.2.0+incompatible // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/gin-gonic/gin v1.8.1 // indirect
	github.com/go-playground/validator/v10 v10.11.1 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rs/cors v1.8.3 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240205150955-31a09d347014 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240213162025-012b6fc9bca9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.6 // indirect
)
