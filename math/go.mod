module cosmossdk.io/math

go 1.22
toolchain go1.24.1

require (
	github.com/cockroachdb/apd/v3 v3.2.1
	github.com/stretchr/testify v1.10.0
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

require (
	cosmossdk.io/errors v1.0.2
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
