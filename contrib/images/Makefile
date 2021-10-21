all: simd-env

simd-env:
	docker build --tag cosmossdk/simd-env -f simd-env/Dockerfile \
		$(shell git rev-parse --show-toplevel)

.PHONY: all simd-env
