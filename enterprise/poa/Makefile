###############################################################################
###                          PoA Module Makefile                            ###
###############################################################################

DOCKER_IMAGE = poa-node
include ../make/enterprise.mk

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
