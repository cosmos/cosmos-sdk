#!/usr/bin/env bash

mockgen_cmd="mockgen"
$mockgen_cmd -package mock -destination ./db_mock.go -source ./types.go