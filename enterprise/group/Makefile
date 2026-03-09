###############################################################################
###                         Group Module Makefile                           ###
###############################################################################

BUILD_MODULE_FIRST = 1
include ../make/enterprise.mk

###############################################################################
###                         Module-Specific Tests                           ###
###############################################################################

.PHONY: test-fuzz

#? test-fuzz: Run types-level fuzz tests (30s each)
test-fuzz:
	@echo "--> Running fuzz tests (group internal math)"
	@go test -fuzz=FuzzNewDecFromString -fuzztime=30s ./x/group/internal/math
	@go test -fuzz=FuzzNewPositiveDecFromString -fuzztime=30s ./x/group/internal/math
	@go test -fuzz=FuzzNewNonNegativeDecFromString -fuzztime=30s ./x/group/internal/math
	@echo "--> Running fuzz tests (group types)"
	@go test -fuzz=FuzzTallyResultGetCounts -fuzztime=30s ./x/group
	@go test -fuzz=FuzzTallyResultTotalCounts -fuzztime=30s ./x/group
	@go test -fuzz=FuzzVoteOptionFromString -fuzztime=30s ./x/group
	@echo "--> Fuzz tests completed"
