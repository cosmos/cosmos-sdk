package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC client sentinel errors
var (
	ErrClientExists                           = sdkerrors.Register(SubModuleName, 2, "light client already exists")
	ErrInvalidClient                          = sdkerrors.Register(SubModuleName, 3, "light client is invalid")
	ErrClientNotFound                         = sdkerrors.Register(SubModuleName, 4, "light client not found")
	ErrClientFrozen                           = sdkerrors.Register(SubModuleName, 5, "light client is frozen due to misbehaviour")
	ErrConsensusStateNotFound                 = sdkerrors.Register(SubModuleName, 6, "consensus state not found")
	ErrInvalidConsensus                       = sdkerrors.Register(SubModuleName, 7, "invalid consensus state")
	ErrClientTypeNotFound                     = sdkerrors.Register(SubModuleName, 8, "client type not found")
	ErrInvalidClientType                      = sdkerrors.Register(SubModuleName, 9, "invalid client type")
	ErrRootNotFound                           = sdkerrors.Register(SubModuleName, 10, "commitment root not found")
	ErrInvalidHeader                          = sdkerrors.Register(SubModuleName, 11, "invalid client header")
	ErrInvalidMisbehaviour                    = sdkerrors.Register(SubModuleName, 12, "invalid light client misbehaviour evidence")
	ErrFailedClientStateVerification          = sdkerrors.Register(SubModuleName, 13, "client state verification failed")
	ErrFailedClientConsensusStateVerification = sdkerrors.Register(SubModuleName, 14, "client consensus state verification failed")
	ErrFailedConnectionStateVerification      = sdkerrors.Register(SubModuleName, 15, "connection state verification failed")
	ErrFailedChannelStateVerification         = sdkerrors.Register(SubModuleName, 16, "channel state verification failed")
	ErrFailedPacketCommitmentVerification     = sdkerrors.Register(SubModuleName, 17, "packet commitment verification failed")
	ErrFailedPacketAckVerification            = sdkerrors.Register(SubModuleName, 18, "packet acknowledgement verification failed")
	ErrFailedPacketAckAbsenceVerification     = sdkerrors.Register(SubModuleName, 19, "packet acknowledgement absence verification failed")
	ErrFailedNextSeqRecvVerification          = sdkerrors.Register(SubModuleName, 20, "next sequence receive verification failed")
	ErrSelfConsensusStateNotFound             = sdkerrors.Register(SubModuleName, 21, "self consensus state not found")
)
