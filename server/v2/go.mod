module cosmossdk.io/server/v2

go 1.23

// server v2 integration (uncomment during development, but comment before release)
// replace (
// 	cosmossdk.io/server/v2/appmanager => ./appmanager
// 	cosmossdk.io/store/v2 => ../../store/v2
// )

require (
	cosmossdk.io/core v1.0.0
	cosmossdk.io/log v1.5.0
	cosmossdk.io/server/v2/appmanager v1.0.0-beta.2
	github.com/cosmos/gogoproto v1.7.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	google.golang.org/genproto/googleapis/api v0.0.0-20241015192408-796eee8c2d53
	google.golang.org/grpc v1.69.4
	google.golang.org/protobuf v1.36.2
)

require (
	cosmossdk.io/schema v1.0.0 // indirect
	github.com/bytedance/sonic v1.12.6 // indirect
	github.com/bytedance/sonic/loader v0.2.1 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.12.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250106144421-5f5ef82da422 // indirect
)
