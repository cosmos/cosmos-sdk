###############################################################################
###                          PoA Module Makefile                            ###
###############################################################################

include ../make/enterprise.mk

###############################################################################
###                         Module-Specific Build                           ###
###############################################################################

.PHONY: build build-docker install

build:
	@go build -C simapp -o ../$(BUILD_DIR)/simd ./simd/

install:
	@go install -C simapp ./simd/

build-docker: build
	@GOOS=linux go build -C simapp -o ../$(BUILD_DIR)/simd-linux ./simd/
	@docker build -t poa-node .

###############################################################################
###                         Module-Specific: Localnet                       ###
###############################################################################

.PHONY: localnet-clean localnet-genesis localnet-distribute localnet-up localnet-down localnet localnet-shell

localnet-clean:
	@./scripts/localnet/clean.sh

localnet-genesis:
	@./scripts/localnet/genesis.sh

localnet-distribute:
	@./scripts/localnet/distribute.sh

localnet-up:
	@docker compose up -d

localnet-down:
	@docker compose down -v

localnet:
	@./scripts/localnet/genesis.sh
	@docker compose up -d
	@docker compose logs node1 -f

localnet-shell:
	@docker compose exec -it node1 bash
