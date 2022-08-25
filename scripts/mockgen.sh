#!/usr/bin/env bash

mockgen_cmd="go run github.com/golang/mock/mockgen"
$mockgen_cmd -source=client/account_retriever.go -package mocks -destination tests/mocks/account_retriever.go
$mockgen_cmd -package mocks -destination tests/mocks/tendermint_tm_db_DB.go github.com/tendermint/tm-db DB
$mockgen_cmd -source db/types.go -package mocks -destination tests/mocks/db/types.go
$mockgen_cmd -source=types/module/module.go -package mocks -destination tests/mocks/types_module_module.go
$mockgen_cmd -source=types/invariant.go -package mocks -destination tests/mocks/types_invariant.go
$mockgen_cmd -source=types/router.go -package mocks -destination tests/mocks/types_router.go
$mockgen_cmd -package mocks -destination tests/mocks/grpc_server.go github.com/gogo/protobuf/grpc Server
$mockgen_cmd -package mocks -destination tests/mocks/tendermint_tendermint_libs_log_DB.go github.com/tendermint/tendermint/libs/log Logger
$mockgen_cmd -source=orm/model/ormtable/hooks.go -package ormmocks -destination orm/testing/ormmocks/hooks.go
$mockgen_cmd -source=x/nft/expected_keepers.go -package testutil -destination x/nft/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/feegrant/expected_keepers.go -package testutil -destination x/feegrant/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/mint/types/expected_keepers.go -package testutil -destination x/mint/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/params/proposal_handler_test.go -package testutil -destination x/params/testutil/staking_keeper_mock.go
$mockgen_cmd -source=x/crisis/types/expected_keepers.go -package testutil -destination x/crisis/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/auth/types/expected_keepers.go -package testutil -destination x/auth/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/auth/ante/expected_keepers.go -package testutil -destination x/auth/ante/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/authz/expected_keepers.go -package testutil -destination x/authz/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/bank/types/expected_keepers.go -package testutil -destination x/bank/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/group/testutil/expected_keepers.go -package testutil -destination x/group/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/evidence/types/expected_keepers.go -package testutil -destination x/evidence/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/slashing/types/expected_keepers.go -package testutil -destination x/slashing/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/genutil/types/expected_keepers.go -package testutil -destination x/genutil/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/gov/testutil/expected_keepers.go -package testutil -destination x/gov/testutil/expected_keepers_mocks.go
