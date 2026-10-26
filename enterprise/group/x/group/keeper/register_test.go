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

package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	group "github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group/keeper"
)

// TestRegisterMsgServerRoutesMsgs registers the group Msg service on a router
// with RegisterMsgServer and runs MsgCreateGroup through its handler.
func (s *TestSuite) TestRegisterMsgServerRoutesMsgs() {
	registry := codectypes.NewInterfaceRegistry()
	group.RegisterInterfaces(registry)
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(registry)
	keeper.RegisterMsgServer(router, s.groupKeeper)

	msg := &group.MsgCreateGroup{
		Admin:   s.addrs[0].String(),
		Members: []group.MemberRequest{{Address: s.addrs[1].String(), Weight: "1"}},
	}
	handler := router.Handler(msg)
	s.Require().NotNil(handler)
	res, err := handler(s.sdkCtx, msg)
	s.Require().NoError(err)
	var created group.MsgCreateGroupResponse
	s.Require().NoError(created.Unmarshal(res.Data))

	info, err := s.groupKeeper.GroupInfo(s.ctx, &group.QueryGroupInfoRequest{GroupId: created.GroupId})
	s.Require().NoError(err)
	s.Require().Equal(s.addrs[0].String(), info.Info.Admin)
}
