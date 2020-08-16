package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC client sentinel errors
var (
	ErrFailUpdateClient                       = sdkerrors.Register(SubModuleName, 1, "unable to update light client")
	ErrClientExists                           = sdkerrors.Register(SubModuleName, 2, "light client already exists")
	ErrClientNotFound                         = sdkerrors.Register(SubModuleName, 3, "light client not found")
	ErrClientFrozen                           = sdkerrors.Register(SubModuleName, 4, "light client is frozen due to misbehaviour")
	ErrConsensusStateNotFound                 = sdkerrors.Register(SubModuleName, 5, "consensus state not found")
	ErrInvalidConsensus                       = sdkerrors.Register(SubModuleName, 6, "invalid consensus state")
	ErrClientTypeNotFound                     = sdkerrors.Register(SubModuleName, 7, "client type not found")
	ErrInvalidClientType                      = sdkerrors.Register(SubModuleName, 8, "invalid client type")
	ErrRootNotFound                           = sdkerrors.Register(SubModuleName, 9, "commitment root not found")
	ErrInvalidHeader                          = sdkerrors.Register(SubModuleName, 10, "invalid client header")
	ErrInvalidEvidence                        = sdkerrors.Register(SubModuleName, 11, "invalid light client misbehaviour evidence")
	ErrFailedClientConsensusStateVerification = sdkerrors.Register(SubModuleName, 12, "client consensus state verification failed")
	ErrFailedConnectionStateVerification      = sdkerrors.Register(SubModuleName, 13, "connection state verification failed")
	ErrFailedChannelStateVerification         = sdkerrors.Register(SubModuleName, 14, "channel state verification failed")
	ErrFailedPacketCommitmentVerification     = sdkerrors.Register(SubModuleName, 15, "packet commitment verification failed")
	ErrFailedPacketAckVerification            = sdkerrors.Register(SubModuleName, 16, "packet acknowledgement verification failed")
	ErrFailedPacketAckAbsenceVerification     = sdkerrors.Register(SubModuleName, 17, "packet acknowledgement absence verification failed")
	ErrFailedNextSeqRecvVerification          = sdkerrors.Register(SubModuleName, 18, "next sequence receive verification failed")
	ErrSelfConsensusStateNotFound             = sdkerrors.Register(SubModuleName, 19, "self consensus state not found")
)
