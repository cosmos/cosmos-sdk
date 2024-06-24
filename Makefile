#!/usr/bin/make -f

include scripts/build/linting.mk
include scripts/build/protobuf.mk
include scripts/build/localnet.mk
include scripts/build/simulations.mk
include scripts/build/testing.mk
include scripts/build/documentation.mk
include scripts/build/build.mk

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy
