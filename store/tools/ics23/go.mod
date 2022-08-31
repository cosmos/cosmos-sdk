module github.com/cosmos/cosmos-sdk/store/tools/ics23

go 1.18

require (
	github.com/confio/ics23/go v0.7.0
	github.com/cosmos/cosmos-db v0.0.0-20220831082506-d49669b17ceb
	github.com/cosmos/cosmos-sdk v0.46.1
	github.com/cosmos/iavl v0.19.2-0.20220831121246-807f8c542e48
	github.com/lazyledger/smt v0.2.1-0.20210709230900-03ea40719554
	github.com/tendermint/tendermint v0.34.21
)

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cockroachdb/errors v1.8.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/pebble v0.0.0-20220817183557-09c6e030a677 // indirect
	github.com/cockroachdb/redact v1.0.8 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.7.5 // indirect
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.20.0 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/sasha-s/go-deadlock v0.2.1-0.20190427202633-1595213edefa // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/sys v0.0.0-20220817070843-5a390386f1f2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/cosmos/cosmos-sdk/store/tools/ics23 => ./
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
)
