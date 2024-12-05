module cosmossdk.io/math

go 1.20

require (
	github.com/cockroachdb/apd/v3 v3.2.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/exp v0.0.0-20240404231335-c0f41cb1a7a0
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240709173604-40e1e62336c5 // indirect
	google.golang.org/grpc v1.64.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

require (
	cosmossdk.io/errors v1.0.1
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	pgregory.net/rapid v1.1.0
)

// Issue with math.Int{}.Size() implementation.
retract [v1.1.0, v1.1.1]

// Bit length differences between Int and Dec
retract (
	v1.3.0
	v1.2.0
	v1.1.2
	[v1.0.0, v1.0.1]
)
