package baseapp

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/types/descriptorpb"
)

func optsFromRaw(raw []byte) *descriptorpb.MessageOptions {
	o := &descriptorpb.MessageOptions{}
	o.ProtoReflect().SetUnknown(raw)
	return o
}

func signerField() []byte {
	b := protowire.AppendTag(nil, 11110000, protowire.BytesType)
	return protowire.AppendString(b, "authority")
}

func internalField(v uint64) []byte {
	b := protowire.AppendTag(nil, internalExtensionNumber, protowire.VarintType)
	return protowire.AppendVarint(b, v)
}

func TestHasInternalOption(t *testing.T) {
	got, err := hasInternalOption(nil)
	require.NoError(t, err)
	require.False(t, got, "nil options must be external")

	got, err = hasInternalOption(optsFromRaw(nil))
	require.NoError(t, err)
	require.False(t, got, "no options must be external")

	got, err = hasInternalOption(optsFromRaw(signerField()))
	require.NoError(t, err)
	require.False(t, got, "other extensions must not mark internal")

	got, err = hasInternalOption(optsFromRaw(internalField(1)))
	require.NoError(t, err)
	require.True(t, got)

	got, err = hasInternalOption(optsFromRaw(internalField(0)))
	require.NoError(t, err)
	require.False(t, got)

	got, err = hasInternalOption(optsFromRaw(append(signerField(), internalField(1)...)))
	require.NoError(t, err)
	require.True(t, got, "internal must be found after other fields")

	// tag without a value must be an error, not a silent "external"
	trunc := protowire.AppendTag(nil, internalExtensionNumber, protowire.VarintType)
	_, err = hasInternalOption(optsFromRaw(trunc))
	require.Error(t, err)
}

func TestMsgServiceRouterIsInternal(t *testing.T) {
	msr := NewMsgServiceRouter()
	msr.internalMsgs["/test.MsgInternal"] = struct{}{}
	require.True(t, msr.IsInternal("/test.MsgInternal"))
	require.False(t, msr.IsInternal("/test.MsgExternal"))
}
