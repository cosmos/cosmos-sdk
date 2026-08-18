module cosmossdk.io/math

go 1.26.6

require (
	github.com/stretchr/testify v1.12.0
	sigs.k8s.io/yaml v1.6.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// reverted the broken Dec type
retract [v1.5.0, v1.5.2]

// Issue with math.Int{}.Size() implementation.
retract [v1.1.0, v1.1.1]

// Bit length differences between Int and Dec
retract (
	v1.3.0
	v1.2.0
	v1.1.2
	[v1.0.0, v1.0.1]
)
