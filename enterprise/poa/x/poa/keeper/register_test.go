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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestRegisterMsgServerRoutesMsgs registers the x/poa Msg service on a router
// with RegisterMsgServer and runs MsgUpdateParams through its handler.
func TestRegisterMsgServerRoutesMsgs(t *testing.T) {
	f := setupTest(t)
	require.NoError(t, f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr}))

	registry := codectypes.NewInterfaceRegistry()
	poatypes.RegisterInterfaces(registry)
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(registry)
	RegisterMsgServer(router, f.poaKeeper)

	newAdmin := sdk.AccAddress("newadmin").String()
	msg := &poatypes.MsgUpdateParams{Admin: adminAddr, Params: poatypes.Params{Admin: newAdmin}}
	handler := router.Handler(msg)
	require.NotNil(t, handler)
	_, err := handler(f.ctx, msg)
	require.NoError(t, err)

	got, err := f.poaKeeper.Params(f.ctx, &poatypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, newAdmin, got.Params.Admin)
}
