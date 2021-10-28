package kvindexer

import (
	"context"
	"errors"
	"github.com/cosmos/cosmos-sdk/orm/apis/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/internal/testkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testproto"
	"github.com/cosmos/cosmos-sdk/orm/pkg/orm"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"testing"
)

func TestIndexer(t *testing.T) {
	ctx := context.TODO()

	indexer := NewIndexer(testkv.New())

	obj := &testproto.PersonStateObject{
		Id:          "name surname",
		City:        "Novi Ligure",
		PostalCode:  15067,
		CountryCode: testproto.CountryCode_IT,
	}

	err := indexer.RegisterObject(ctx, &v1alpha1.StateObjectDescriptor{
		TypePrefix:        []byte{0x0, 0x1},
		TableDescriptor:   proto.GetExtension(obj.ProtoReflect().Descriptor().Options(), v1alpha1.E_TableDescriptor).(*v1alpha1.TableDescriptor),
		FileDescriptor:    nil,
		ProtoDependencies: nil,
	}, obj.ProtoReflect().Type())

	require.NoError(t, err)

	require.NoError(t, indexer.Index(ctx, []byte(obj.Id), obj))

	iter, err := indexer.List(ctx, obj, orm.ListOptions{
		FieldsToMatch: []orm.FieldMatch{
			{
				FieldName: "postal_code",
				Value:     protoreflect.ValueOfInt64(15067),
			},
		},
	})

	require.NoError(t, err)
	require.True(t, iter.Valid())
	require.Equal(t, obj.Id, string(iter.Key()))
	iter.Close()

	require.NoError(t, indexer.ClearIndexes(ctx, []byte(obj.Id), obj))

	iter, err = indexer.List(ctx, obj, orm.ListOptions{
		FieldsToMatch: []orm.FieldMatch{
			{
				FieldName: "postal_code",
				Value:     protoreflect.ValueOfInt64(15067),
			},
		},
	})

	require.True(t, errors.Is(err, orm.ErrNotFound))
}
