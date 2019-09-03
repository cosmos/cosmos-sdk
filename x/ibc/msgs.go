package ibc

import (
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

type MsgCreateClient = client.MsgCreateClient
type MsgUpdateClient = client.MsgUpdateClient
type MsgOpenInitConnection = connection.MsgOpenInit
type MsgOpenTryConnection = connection.MsgOpenTry
type MsgOpenAckConnection = connection.MsgOpenAck
type MsgOpenConfirmConnection = connection.MsgOpenConfirm
type MsgOpenInitChannel = channel.MsgOpenInit
type MsgOpenTryChannel = channel.MsgOpenTry
type MsgOpenAckChannel = channel.MsgOpenAck
type MsgOpenConfirmChannel = channel.MsgOpenConfirm
type MsgPacket = channel.MsgPacket
