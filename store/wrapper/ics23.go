package wrapper

import (
	ics23 "github.com/confio/ics23/go"
	cics23 "github.com/cosmos/ics23/go"
	"github.com/golang/protobuf/proto"
)

func ConvertCommitmentProof(proof *cics23.CommitmentProof) *ics23.CommitmentProof {
	msg, _ := proto.Marshal(proof)
	var newProof ics23.CommitmentProof
	proto.Unmarshal(msg, &newProof)
	return &newProof
}
