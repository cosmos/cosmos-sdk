package v040

import (
	v039slashing "github.com/cosmos/cosmos-sdk/x/slashing/legacy/v0_39"
)

// DONTCOVER
// nolint

const (
	ModuleName = "slashing"
)

// SigningInfo stores validator signing info of corresponding address
type SigningInfo struct {
	Address              string                            `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	ValidatorSigningInfo v039slashing.ValidatorSigningInfo `protobuf:"bytes,2,opt,name=validator_signing_info,json=validatorSigningInfo,proto3" json:"validator_signing_info" yaml:"validator_signing_info"`
}

// ValidatorMissedBlocks contains array of missed blocks of corresponding address
type ValidatorMissedBlocks struct {
	Address      string                     `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	MissedBlocks []v039slashing.MissedBlock `protobuf:"bytes,2,rep,name=missed_blocks,json=missedBlocks,proto3" json:"missed_blocks" yaml:"missed_blocks"`
}

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params       v039slashing.Params     `protobuf:"bytes,1,opt,name=params,proto3,casttype=Params" json:"params"`
	SigningInfos []SigningInfo           `protobuf:"bytes,2,rep,name=signing_infos,json=signingInfos,proto3" json:"signing_infos" yaml:"signing_infos"`
	MissedBlocks []ValidatorMissedBlocks `protobuf:"bytes,3,rep,name=missed_blocks,json=missedBlocks,proto3" json:"missed_blocks" yaml:"missed_blocks"`
}
