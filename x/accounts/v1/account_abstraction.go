package v1

import (
	accountsv1 "cosmossdk.io/api/cosmos/accounts/v1"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"google.golang.org/protobuf/types/known/anypb"
)

func GogoUserOpToProtoV2(op *UserOperation) *accountsv1.UserOperation {
	return &accountsv1.UserOperation{
		Sender:                 op.Sender,
		AuthenticationMethod:   op.AuthenticationMethod,
		AuthenticationData:     op.AuthenticationData,
		Sequence:               op.Sequence,
		AuthenticationGasLimit: op.AuthenticationGasLimit,
		BundlerPaymentMessages: intoProtoV2Any(op.BundlerPaymentMessages),
		BundlerPaymentGasLimit: op.BundlerPaymentGasLimit,
		ExecutionMessages:      intoProtoV2Any(op.ExecutionMessages),
		ExecutionGasLimit:      op.ExecutionGasLimit,
	}
}

func intoProtoV2Any(msgs []*codectypes.Any) []*anypb.Any {
	protoMsgs := make([]*anypb.Any, len(msgs))
	for i, msg := range msgs {
		protoMsgs[i] = &anypb.Any{
			TypeUrl: msg.TypeUrl,
			Value:   msg.Value,
		}
	}
	return protoMsgs
}
