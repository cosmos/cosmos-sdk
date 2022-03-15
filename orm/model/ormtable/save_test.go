package ormtable_test

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"testing"

	"github.com/regen-network/gocuke"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
	"github.com/cosmos/cosmos-sdk/orm/testing/ormtest"
)

func TestSave(t *testing.T) {
	gocuke.NewRunner(t, &suite{}).Path("../../features/table/save.feature").Run()
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
	assert.ErrorIs(s, s.err, s.toError(a), s.err.Error())
}

func (s *suite) toError(str string) error {
	switch str {
	case "already exists":
		return ormerrors.AlreadyExists
	case "not found":
		return ormerrors.NotFound
	case "constraint violation":
		return ormerrors.ConstraintViolation
	case "unique key violation":
		return ormerrors.UniqueKeyViolation
	default:
		s.Fatalf("missing case for error %s", str)
		return nil
	}
}

func (s *suite) ExpectGrpcErrorCode(a string) {
	var code codes.Code
	assert.NilError(s, code.UnmarshalJSON([]byte(fmt.Sprintf("%q", a))))
	assert.Equal(s, code, status.Code(s.err))
}
