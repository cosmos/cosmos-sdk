#!/usr/bin/make -f

########################################
### Simulations

runsim: $(GOBIN)/runsim
$(GOBIN)/runsim: contrib/runsim/main.go
	go install github.com/cosmos/cosmos-sdk/cmd/gaia/contrib/runsim

sim-gaia-nondeterminism:
	@echo "Running nondeterminism test..."
	@go test -mod=readonly ./cmd/gaia/app -run TestAppStateDeterminism -SimulationEnabled=true -v -timeout 10m

sim-gaia-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	@go test -mod=readonly github.com/cosmos/cosmos-sdk/cmd/gaia/app -run TestFullGaiaSimulation -SimulationGenesis=${HOME}/.gaiad/config/genesis.json \
		-SimulationEnabled=true -SimulationNumBlocks=100 -SimulationBlockSize=200 -SimulationCommit=true -SimulationSeed=99 -SimulationPeriod=5 -v -timeout 24h

sim-gaia-fast:
	@echo "Running quick Gaia simulation. This may take several minutes..."
	@go test -mod=readonly github.com/cosmos/cosmos-sdk/cmd/gaia/app -run TestFullGaiaSimulation -SimulationEnabled=true -SimulationNumBlocks=100 -SimulationBlockSize=200 -SimulationCommit=true -SimulationSeed=99 -SimulationPeriod=5 -v -timeout 24h

sim-gaia-import-export: runsim
	@echo "Running Gaia import/export simulation. This may take several minutes..."
	$(GOBIN)/runsim 50 5 TestGaiaImportExport

sim-gaia-simulation-after-import: runsim
	@echo "Running Gaia simulation-after-import. This may take several minutes..."
	$(GOBIN)/runsim 50 5 TestGaiaSimulationAfterImport

sim-gaia-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	$(GOBIN)/runsim -g ${HOME}/.gaiad/config/genesis.json 400 5 TestFullGaiaSimulation

sim-gaia-multi-seed: runsim
	@echo "Running multi-seed Gaia simulation. This may take awhile!"
	$(GOBIN)/runsim 400 5 TestFullGaiaSimulation

sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	@go test -mod=readonly github.com/cosmos/cosmos-sdk/cmd/gaia/app -benchmem -bench=BenchmarkInvariants -run=^$ \
	-SimulationEnabled=true -SimulationNumBlocks=1000 -SimulationBlockSize=200 \
	-SimulationCommit=true -SimulationSeed=57 -v -timeout 24h

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true
sim-gaia-benchmark:
	@echo "Running Gaia benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ github.com/cosmos/cosmos-sdk/cmd/gaia/app -bench ^BenchmarkFullGaiaSimulation$$  \
		-SimulationEnabled=true -SimulationNumBlocks=$(SIM_NUM_BLOCKS) -SimulationBlockSize=$(SIM_BLOCK_SIZE) -SimulationCommit=$(SIM_COMMIT) -timeout 24h

sim-gaia-profile:
	@echo "Running Gaia benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ github.com/cosmos/cosmos-sdk/cmd/gaia/app -bench ^BenchmarkFullGaiaSimulation$$ \
		-SimulationEnabled=true -SimulationNumBlocks=$(SIM_NUM_BLOCKS) -SimulationBlockSize=$(SIM_BLOCK_SIZE) -SimulationCommit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

.PHONY: sim-gaia-nondeterminism sim-gaia-custom-genesis-fast sim-gaia-fast sim-gaia-import-export \
	sim-gaia-simulation-after-import sim-gaia-custom-genesis-multi-seed sim-gaia-multi-seed \
	sim-benchmark-invariants sim-gaia-benchmark sim-gaia-profile
