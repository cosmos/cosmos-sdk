package ormmocks

import (
	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

// Code adapted from MIT-licensed https://github.com/budougumi0617/cmpmock/blob/master/diffmatcher.go

// Eq returns a gomock.Matcher which uses go-cmp to compare protobuf messages.
func Eq(message proto.Message) gomock.Matcher {
	return &protoEq{message: message}
}

type protoEq struct {
	message interface{}
	diff    string
}

func (p protoEq) Matches(x interface{}) bool {
	p.diff = cmp.Diff(x, p.message, protocmp.Transform())
	return len(p.diff) == 0
}

func (p protoEq) String() string {
	return p.diff
}
