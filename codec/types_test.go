package codec_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testdata"
)

func createTestCodec() *amino.Codec {
	cdc := amino.NewCodec()

	cdc.RegisterInterface((*testdata.Animal)(nil), nil)
	cdc.RegisterConcrete(testdata.Dog{}, "testdata/Dog", nil)
	cdc.RegisterConcrete(testdata.Cat{}, "testdata/Cat", nil)

	return cdc
}

func TestBaseCodec_MarshalBinaryBare(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := testdata.Dog{Name: "Rufus"}

	bz, err := cdc.MarshalBinaryBare(&dog)
	require.NoError(t, err)
	exp, err := dog.Marshal()
	require.NoError(t, err)
	require.Equal(t, exp, bz)

	cat := testdata.Cat{Name: "Lisa"}

	bz, err = cdc.MarshalBinaryBare(&cat)
	require.NoError(t, err)
	exp, err = aminoCdc.MarshalBinaryBare(&cat)
	require.NoError(t, err)
	require.Equal(t, exp, bz)
}

func TestBaseCodec_MustMarshalBinaryBare(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := testdata.Dog{Name: "Rufus"}
	var bz []byte

	require.NotPanics(t, func() { bz = cdc.MustMarshalBinaryBare(&dog) })
	exp, err := dog.Marshal()
	require.NoError(t, err)
	require.Equal(t, exp, bz)

	cat := testdata.Cat{Name: "Lisa"}

	require.NotPanics(t, func() { bz = cdc.MustMarshalBinaryBare(&cat) })
	exp, err = aminoCdc.MarshalBinaryBare(&cat)
	require.NoError(t, err)
	require.Equal(t, exp, bz)
}

func TestBaseCodec_MarshalBinaryLengthPrefixed(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := testdata.Dog{Name: "Rufus"}

	bz, err := cdc.MarshalBinaryLengthPrefixed(&dog)
	require.NoError(t, err)
	exp, err := dog.Marshal()
	require.NoError(t, err)
	require.Equal(t, int(bz[:1][0]), dog.Size())
	require.Equal(t, exp, bz[1:])

	cat := testdata.Cat{Name: "Lisa"}

	bz, err = cdc.MarshalBinaryLengthPrefixed(&cat)
	require.NoError(t, err)
	exp, err = aminoCdc.MarshalBinaryLengthPrefixed(&cat)
	require.NoError(t, err)
	require.Equal(t, exp, bz)
}

func TestBaseCodec_MustMarshalBinaryLengthPrefixed(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := testdata.Dog{Name: "Rufus"}

	var bz []byte
	require.NotPanics(t, func() { bz = cdc.MustMarshalBinaryLengthPrefixed(&dog) })
	exp, err := dog.Marshal()
	require.NoError(t, err)
	require.Equal(t, int(bz[:1][0]), dog.Size())
	require.Equal(t, exp, bz[1:])

	cat := testdata.Cat{Name: "Lisa"}

	require.NotPanics(t, func() { bz = cdc.MustMarshalBinaryLengthPrefixed(&cat) })
	exp, err = aminoCdc.MarshalBinaryLengthPrefixed(&cat)
	require.NoError(t, err)
	require.Equal(t, exp, bz)
}

func TestBaseCodec_UnmarshalBinaryBare(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := &testdata.Dog{Name: "Rufus"}
	bz, err := cdc.MarshalBinaryBare(dog)
	require.NoError(t, err)

	otherDog := new(testdata.Dog)
	require.NoError(t, cdc.UnmarshalBinaryBare(bz, otherDog))
	require.Equal(t, dog, otherDog)

	cat := &testdata.Cat{Name: "Lisa"}
	bz, err = cdc.MarshalBinaryBare(cat)
	require.NoError(t, err)

	otherCat := new(testdata.Cat)
	require.NoError(t, aminoCdc.UnmarshalBinaryBare(bz, otherCat))
	require.Equal(t, cat, otherCat)
}

func TestBaseCodec_MustUnmarshalBinaryBare(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := &testdata.Dog{Name: "Rufus"}
	bz, err := cdc.MarshalBinaryBare(dog)
	require.NoError(t, err)

	otherDog := new(testdata.Dog)
	require.NotPanics(t, func() { cdc.MustUnmarshalBinaryBare(bz, otherDog) })
	require.Equal(t, dog, otherDog)

	cat := &testdata.Cat{Name: "Lisa"}
	bz, err = cdc.MarshalBinaryBare(cat)
	require.NoError(t, err)

	otherCat := new(testdata.Cat)
	require.NotPanics(t, func() { aminoCdc.MustUnmarshalBinaryBare(bz, otherCat) })
	require.Equal(t, cat, otherCat)
}

func TestBaseCodec_UnmarshalBinaryLengthPrefixed(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := &testdata.Dog{Name: "Rufus"}
	bz, err := cdc.MarshalBinaryLengthPrefixed(dog)
	require.NoError(t, err)

	otherDog := new(testdata.Dog)
	require.NoError(t, cdc.UnmarshalBinaryLengthPrefixed(bz, otherDog))
	require.Equal(t, dog, otherDog)

	cat := &testdata.Cat{Name: "Lisa"}
	bz, err = cdc.MarshalBinaryLengthPrefixed(cat)
	require.NoError(t, err)

	otherCat := new(testdata.Cat)
	require.NoError(t, aminoCdc.UnmarshalBinaryLengthPrefixed(bz, otherCat))
	require.Equal(t, cat, otherCat)
}

func TestBaseCodec_UnmarshalBinaryLengthPrefixedInvalidPrefix(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := &testdata.Dog{Name: "Rufus"}
	bz, err := cdc.MarshalBinaryLengthPrefixed(dog)
	require.NoError(t, err)

	otherDog := new(testdata.Dog)
	bz[0] = 50
	require.Error(t, cdc.UnmarshalBinaryLengthPrefixed(bz, otherDog), "expected error for not enough bytes to read")
	bz[0] = 1
	require.Error(t, cdc.UnmarshalBinaryLengthPrefixed(bz, otherDog), "expected error for too many bytes to read")
}

func TestBaseCodec_MustUnmarshalBinaryLengthPrefixed(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := &testdata.Dog{Name: "Rufus"}
	bz, err := cdc.MarshalBinaryLengthPrefixed(dog)
	require.NoError(t, err)

	otherDog := new(testdata.Dog)
	require.NotPanics(t, func() { cdc.MustUnmarshalBinaryLengthPrefixed(bz, otherDog) })
	require.Equal(t, dog, otherDog)

	cat := &testdata.Cat{Name: "Lisa"}
	bz, err = cdc.MarshalBinaryLengthPrefixed(cat)
	require.NoError(t, err)

	otherCat := new(testdata.Cat)
	require.NotPanics(t, func() { aminoCdc.MustUnmarshalBinaryLengthPrefixed(bz, otherCat) })
	require.Equal(t, cat, otherCat)
}

func TestBaseCodec_NoAminoCodec(t *testing.T) {
	cdc := codec.NewBaseCodec(nil)
	cat := &testdata.Cat{Name: "Lisa"}

	_, err := cdc.MarshalBinaryBare(cat)
	require.Error(t, err)
	require.Panics(t, func() { cdc.MustMarshalBinaryBare(cat) })

	_, err = cdc.MarshalBinaryLengthPrefixed(cat)
	require.Error(t, err)
	require.Panics(t, func() { cdc.MustMarshalBinaryLengthPrefixed(cat) })
}

func TestBaseCodec_MarshalJSON(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := testdata.Dog{Name: "Rufus"}

	bz, err := cdc.MarshalJSON(&dog)
	require.NoError(t, err)
	require.Equal(t, "{\"name\":\"Rufus\"}", string(bz))

	cat := testdata.Cat{Name: "Lisa"}

	bz, err = cdc.MarshalJSON(&cat)
	require.NoError(t, err)
	require.Equal(t, "{\"type\":\"testdata/Cat\",\"value\":{\"Name\":\"Lisa\"}}", string(bz))
}

func TestBaseCodec_MustMarshalJSON(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := testdata.Dog{Name: "Rufus"}

	var bz []byte
	require.NotPanics(t, func() { bz = cdc.MustMarshalJSON(&dog) })
	require.Equal(t, "{\"name\":\"Rufus\"}", string(bz))

	cat := testdata.Cat{Name: "Lisa"}

	require.NotPanics(t, func() { bz = cdc.MustMarshalJSON(&cat) })
	require.Equal(t, "{\"type\":\"testdata/Cat\",\"value\":{\"Name\":\"Lisa\"}}", string(bz))
}

func TestBaseCodec_UnmarshalJSON(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := &testdata.Dog{Name: "Rufus"}
	bz, err := cdc.MarshalJSON(dog)
	require.NoError(t, err)

	otherDog := new(testdata.Dog)
	require.NoError(t, cdc.UnmarshalJSON(bz, otherDog))
	require.Equal(t, dog, otherDog)

	cat := &testdata.Cat{Name: "Lisa"}
	bz, err = cdc.MarshalJSON(cat)
	require.NoError(t, err)

	otherCat := new(testdata.Cat)
	require.NoError(t, aminoCdc.UnmarshalJSON(bz, otherCat))
	require.Equal(t, cat, otherCat)
}

func TestBaseCodec_MustUnmarshalJSON(t *testing.T) {
	aminoCdc := createTestCodec()
	cdc := codec.NewBaseCodec(aminoCdc)

	dog := &testdata.Dog{Name: "Rufus"}
	bz, err := cdc.MarshalJSON(dog)
	require.NoError(t, err)

	otherDog := new(testdata.Dog)
	require.NotPanics(t, func() { cdc.UnmarshalJSON(bz, otherDog) })
	require.Equal(t, dog, otherDog)

	cat := &testdata.Cat{Name: "Lisa"}
	bz, err = cdc.MarshalJSON(cat)
	require.NoError(t, err)

	otherCat := new(testdata.Cat)
	require.NotPanics(t, func() { aminoCdc.UnmarshalJSON(bz, otherCat) })
	require.Equal(t, cat, otherCat)
}
