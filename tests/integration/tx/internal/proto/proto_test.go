package proto

import (
	v1 "cosmossdk.io/x/gov/types/v1"
	"fmt"
	"google.golang.org/protobuf/protoadapt"
	"testing"
)

func Test_ProtoStuff(t *testing.T) {
	m := &v1.MsgSubmitProposal{
		Messages:       nil,
		InitialDeposit: nil,
		Proposer:       "",
		Metadata:       "",
		Title:          "title",
		Summary:        "",
		Expedited:      false,
		ProposalType:   0,
	}
	fmt.Println(m.String())

	mv2 := protoadapt.MessageV2Of(m)
	fmt.Println(mv2)
	desc := mv2.ProtoReflect().Descriptor()
	for i := 0; i < desc.Fields().Len(); i++ {
		field := desc.Fields().Get(i)
		fmt.Println(field.Name())
	}
}
