module github.com/cosmos/cosmos-sdk/app

go 1.17

require (
	github.com/cosmos/cosmos-sdk/api v0.1.0-alpha4
	github.com/cosmos/cosmos-sdk/container v1.0.0-alpha.2
	google.golang.org/protobuf v1.27.1
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/cosmos/cosmos-proto v1.0.0-alpha7 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/goccy/go-graphviz v0.0.9 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/image v0.0.0-20200119044424-58c23975cae1 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb // indirect
	google.golang.org/grpc v1.44.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/cosmos/cosmos-sdk/api => ../api

replace github.com/cosmos/cosmos-sdk/container => ../container
