all: simd-env

simd-env:
	docker build --build-arg UID=$(shell id -u) --build-arg GID=$(shell id -g) --tag cosmossdk/simd-env simd-env

rbuilder:
	docker build --tag cosmossdk/rbuilder rbuilder

.PHONY: all simd-env rbuilder
