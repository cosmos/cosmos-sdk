// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/group/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

/*
Package codec provides a singleton instance of Amino codec that should be used to register
any concrete type that can later be referenced inside a MsgSubmitProposal instance so that they
can be (de)serialized properly.

Amino types should be ideally registered inside this codec within the init function of each module's
codec.go file as follows:

	func init() {
		// ...

		RegisterLegacyAminoCodec(govcodec.Amino)
	RegisterLegacyAminoCodec(groupcodec.Amino)

	}

The codec instance is put inside this package and not the x/gov/types package in order to avoid any dependency cycle.
*/
package codec
