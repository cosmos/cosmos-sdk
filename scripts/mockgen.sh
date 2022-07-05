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