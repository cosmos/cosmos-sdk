GO_RUN= go run -mod=readonly
GENDIR ?= ./testdata

.PHONY: testgen testgen-iavl testgen-smt testgen-simple

# make sure we turn on go modules
export GO111MODULE := on

# Usage: GENDIR=/path/to/ics23/testdata/iavl make testgen-iavl

testgen: testgen-iavl testgen-smt testgen-simple

testgen-iavl:
	@mkdir -p "$(GENDIR)"
	$(GO_RUN) ./iavl/cmd/testgen-iavl exist left 987 > "$(GENDIR)"/exist_left.json
	$(GO_RUN) ./iavl/cmd/testgen-iavl exist middle 812 > "$(GENDIR)"/exist_middle.json
	$(GO_RUN) ./iavl/cmd/testgen-iavl exist right 1261 > "$(GENDIR)"/exist_right.json
	$(GO_RUN) ./iavl/cmd/testgen-iavl nonexist left 813 > "$(GENDIR)"/nonexist_left.json
	$(GO_RUN) ./iavl/cmd/testgen-iavl nonexist middle 691 > "$(GENDIR)"/nonexist_middle.json
	$(GO_RUN) ./iavl/cmd/testgen-iavl nonexist right 1535 > "$(GENDIR)"/nonexist_right.json
	$(GO_RUN) ./iavl/cmd/testgen-iavl batch 1801 20 0 > "$(GENDIR)"/batch_exist.json
	$(GO_RUN) ./iavl/cmd/testgen-iavl batch 1807 0 20 > "$(GENDIR)"/batch_nonexist.json

testgen-smt:
	@mkdir -p "$(GENDIR)"
	$(GO_RUN) ./smt/cmd/testgen-smt exist left 987 > "$(GENDIR)"/exist_left.json
	$(GO_RUN) ./smt/cmd/testgen-smt exist middle 812 > "$(GENDIR)"/exist_middle.json
	$(GO_RUN) ./smt/cmd/testgen-smt exist right 1261 > "$(GENDIR)"/exist_right.json
	$(GO_RUN) ./smt/cmd/testgen-smt nonexist left 813 > "$(GENDIR)"/nonexist_left.json
	$(GO_RUN) ./smt/cmd/testgen-smt nonexist middle 691 > "$(GENDIR)"/nonexist_middle.json
	$(GO_RUN) ./smt/cmd/testgen-smt nonexist right 1535 > "$(GENDIR)"/nonexist_right.json
	$(GO_RUN) ./smt/cmd/testgen-smt batch 1801 20 0 > "$(GENDIR)"/batch_exist.json
	$(GO_RUN) ./smt/cmd/testgen-smt batch 1807 0 20 > "$(GENDIR)"/batch_nonexist.json

testgen-simple:
	@mkdir -p "$(GENDIR)"
	$(GO_RUN) ./tendermint/cmd/testgen-simple exist left 987 > "$(GENDIR)"/exist_left.json
	$(GO_RUN) ./tendermint/cmd/testgen-simple exist middle 812 > "$(GENDIR)"/exist_middle.json
	$(GO_RUN) ./tendermint/cmd/testgen-simple exist right 1261 > "$(GENDIR)"/exist_right.json
	$(GO_RUN) ./tendermint/cmd/testgen-simple nonexist left 813 > "$(GENDIR)"/nonexist_left.json
	$(GO_RUN) ./tendermint/cmd/testgen-simple nonexist middle 691 > "$(GENDIR)"/nonexist_middle.json
	$(GO_RUN) ./tendermint/cmd/testgen-simple nonexist right 1535 > "$(GENDIR)"/nonexist_right.json
	$(GO_RUN) ./tendermint/cmd/testgen-simple batch 1801 20 0 > "$(GENDIR)"/batch_exist.json
	$(GO_RUN) ./tendermint/cmd/testgen-simple batch 1807 0 20 > "$(GENDIR)"/batch_nonexist.json
