package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC client sentinel errors
var (
	ErrClientExists                           = sdkerrors.Register(SubModuleName, 1, "light client already exists")
	ErrClientNotFound                         = sdkerrors.Register(SubModuleName, 2, "light client not found")
	ErrClientFrozen                           = sdkerrors.Register(SubModuleName, 3, "light client is frozen due to misbehaviour")
	ErrConsensusStateNotFound                 = sdkerrors.Register(SubModuleName, 4, "consensus state not found")
	ErrInvalidConsensus                       = sdkerrors.Register(SubModuleName, 5, "invalid consensus state")
	ErrClientTypeNotFound                     = sdkerrors.Register(SubModuleName, 6, "client type not found")
	ErrInvalidClientType                      = sdkerrors.Register(SubModuleName, 7, "invalid client type")
	ErrRootNotFound                           = sdkerrors.Register(SubModuleName, 8, "commitment root not found")
	ErrInvalidHeader                          = sdkerrors.Register(SubModuleName, 9, "invalid block header")
	ErrInvalidEvidence                        = sdkerrors.Register(SubModuleName, 10, "invalid light client misbehaviour evidence")
	ErrFailedClientConsensusStateVerification = sdkerrors.Register(SubModuleName, 13, "client consensus state verification failed")
	ErrFailedConnectionStateVerification      = sdkerrors.Register(SubModuleName, 14, "connection state verification failed")
	ErrFailedChannelStateVerification         = sdkerrors.Register(SubModuleName, 15, "channel state verification failed")
	ErrFailedPacketCommitmentVerification     = sdkerrors.Register(SubModuleName, 16, "packet commitment verification failed")
	ErrFailedPacketAckVerification            = sdkerrors.Register(SubModuleName, 17, "packet acknowledgement verification failed")
	ErrFailedPacketAckAbsenceVerification     = sdkerrors.Register(SubModuleName, 18, "packet acknowledgement absence verification failed")
	ErrFailedNextSeqRecvVerification          = sdkerrors.Register(SubModuleName, 19, "next sequence receive verification failed")
	ErrSelfConsensusStateNotFound             = sdkerrors.Register(SubModuleName, 20, "self consensus state not found")
)
