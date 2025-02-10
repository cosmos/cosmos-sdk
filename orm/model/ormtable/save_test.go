package ormtable_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/regen-network/gocuke"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/internal/testpb"
	"cosmossdk.io/orm/model/ormtable"
	"cosmossdk.io/orm/testing/ormtest"
)

func TestSave(t *testing.T) {
	gocuke.NewRunner(t, &suite{}).Path("../../features/table/saving.feature").Run()
}

type suite struct {
	gocuke.TestingT
	table ormtable.Table
	ctx   context.Context
	err   error
}

func (s *suite) Before() {
	var err error
	s.table, err = ormtable.Build(ormtable.Options{
		MessageType: (&testpb.SimpleExample{}).ProtoReflect().Type(),
	})
	assert.NilError(s, err)
	s.ctx = ormtable.WrapContextDefault(ormtest.NewMemoryBackend())
}

func (s *suite) AnExistingEntity(docString gocuke.DocString) {
	existing := s.simpleExampleFromDocString(docString)
	assert.NilError(s, s.table.Insert(s.ctx, existing))
}

func (s suite) simpleExampleFromDocString(docString gocuke.DocString) *testpb.SimpleExample {
	ex := &testpb.SimpleExample{}
	assert.NilError(s, protojson.Unmarshal([]byte(docString.Content), ex))
	return ex
}

func (s *suite) IInsert(a gocuke.DocString) {
	ex := s.simpleExampleFromDocString(a)
	s.err = s.table.Insert(s.ctx, ex)
}

func (s *suite) IUpdate(a gocuke.DocString) {
	ex := s.simpleExampleFromDocString(a)
	s.err = s.table.Update(s.ctx, ex)
}

func (s *suite) ExpectAError(a string) {
	assert.ErrorContains(s, s.err, a)
}

func (s *suite) ExpectGrpcErrorCode(a string) {
	var code codes.Code
	assert.NilError(s, code.UnmarshalJSON([]byte(fmt.Sprintf("%q", a))))
	assert.Equal(s, code, status.Code(s.err))
}
