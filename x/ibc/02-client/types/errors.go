package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC client sentinel errors
var (
	ErrClientExists                           = sdkerrors.Register(SubModuleName, 2, "light client already exists")
<<<<<<< HEAD
	ErrClientNotFound                         = sdkerrors.Register(SubModuleName, 3, "light client not found")
	ErrClientFrozen                           = sdkerrors.Register(SubModuleName, 4, "light client is frozen due to misbehaviour")
	ErrConsensusStateNotFound                 = sdkerrors.Register(SubModuleName, 5, "consensus state not found")
	ErrInvalidConsensus                       = sdkerrors.Register(SubModuleName, 6, "invalid consensus state")
	ErrClientTypeNotFound                     = sdkerrors.Register(SubModuleName, 7, "client type not found")
	ErrInvalidClientType                      = sdkerrors.Register(SubModuleName, 8, "invalid client type")
	ErrRootNotFound                           = sdkerrors.Register(SubModuleName, 9, "commitment root not found")
	ErrInvalidHeight                          = sdkerrors.Register(SubModuleName, 10, "invalid height")
	ErrInvalidHeader                          = sdkerrors.Register(SubModuleName, 11, "invalid client header")
	ErrInvalidEvidence                        = sdkerrors.Register(SubModuleName, 12, "invalid light client misbehaviour evidence")
	ErrFailedClientConsensusStateVerification = sdkerrors.Register(SubModuleName, 13, "client consensus state verification failed")
	ErrFailedConnectionStateVerification      = sdkerrors.Register(SubModuleName, 14, "connection state verification failed")
	ErrFailedChannelStateVerification         = sdkerrors.Register(SubModuleName, 15, "channel state verification failed")
	ErrFailedPacketCommitmentVerification     = sdkerrors.Register(SubModuleName, 16, "packet commitment verification failed")
	ErrFailedPacketAckVerification            = sdkerrors.Register(SubModuleName, 17, "packet acknowledgement verification failed")
	ErrFailedPacketAckAbsenceVerification     = sdkerrors.Register(SubModuleName, 18, "packet acknowledgement absence verification failed")
	ErrFailedNextSeqRecvVerification          = sdkerrors.Register(SubModuleName, 19, "next sequence receive verification failed")
	ErrSelfConsensusStateNotFound             = sdkerrors.Register(SubModuleName, 20, "self consensus state not found")
=======
	ErrInvalidClient                          = sdkerrors.Register(SubModuleName, 3, "light client is invalid")
	ErrClientNotFound                         = sdkerrors.Register(SubModuleName, 4, "light client not found")
	ErrClientFrozen                           = sdkerrors.Register(SubModuleName, 5, "light client is frozen due to misbehaviour")
	ErrConsensusStateNotFound                 = sdkerrors.Register(SubModuleName, 6, "consensus state not found")
	ErrInvalidConsensus                       = sdkerrors.Register(SubModuleName, 7, "invalid consensus state")
	ErrClientTypeNotFound                     = sdkerrors.Register(SubModuleName, 8, "client type not found")
	ErrInvalidClientType                      = sdkerrors.Register(SubModuleName, 9, "invalid client type")
	ErrRootNotFound                           = sdkerrors.Register(SubModuleName, 10, "commitment root not found")
	ErrInvalidHeader                          = sdkerrors.Register(SubModuleName, 11, "invalid client header")
	ErrInvalidEvidence                        = sdkerrors.Register(SubModuleName, 12, "invalid light client misbehaviour evidence")
	ErrFailedClientStateVerification          = sdkerrors.Register(SubModuleName, 13, "client state verification failed")
	ErrFailedClientConsensusStateVerification = sdkerrors.Register(SubModuleName, 14, "client consensus state verification failed")
	ErrFailedConnectionStateVerification      = sdkerrors.Register(SubModuleName, 15, "connection state verification failed")
	ErrFailedChannelStateVerification         = sdkerrors.Register(SubModuleName, 16, "channel state verification failed")
	ErrFailedPacketCommitmentVerification     = sdkerrors.Register(SubModuleName, 17, "packet commitment verification failed")
	ErrFailedPacketAckVerification            = sdkerrors.Register(SubModuleName, 18, "packet acknowledgement verification failed")
	ErrFailedPacketAckAbsenceVerification     = sdkerrors.Register(SubModuleName, 19, "packet acknowledgement absence verification failed")
	ErrFailedNextSeqRecvVerification          = sdkerrors.Register(SubModuleName, 20, "next sequence receive verification failed")
	ErrSelfConsensusStateNotFound             = sdkerrors.Register(SubModuleName, 21, "self consensus state not found")
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74
)
