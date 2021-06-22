module github.com/cosmos/cosmos-sdk/cosmovisor

go 1.14

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/hashicorp/go-getter v1.4.1
	github.com/otiai10/copy v1.4.2
	github.com/stretchr/testify v1.7.0
)

// replace github.com/cosmos/cosmos-sdk => ../

// replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
// replace google.golang.org/grpc => google.golang.org/grpc v1.33.2
