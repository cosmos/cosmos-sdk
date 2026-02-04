package types_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/store/types"
)

func TestBytesIsZero(t *testing.T) {
	tests := []struct {
		tc       []byte
		expected bool
	}{
		{[]byte("foobar"), false},
		{[]byte(""), false},
		{[]byte{}, false},
		{nil, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			assert.Equal(t, test.expected, types.BytesIsZero(test.tc))
		})
	}
}

func TestBytesValueLen(t *testing.T) {
	tests := []struct {
		tc       []byte
		expected int
	}{
		{[]byte("foobar"), 6},
		{[]byte(""), 0},
		{[]byte{}, 0},
		{nil, 0},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, types.BytesValueLen(test.tc))
	}
}

type struct1 struct {
	a int
	b []string
}

type struct2 struct {
	a int
	b []string
	c bool
}

func TestAnyIsZero(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		obj      any
		expected bool
	}{
		{
			"string length",
			"foobar",
			false,
		},
		{
			"struct1 length",
			struct1{4, []string{"1", "2", "3"}},
			false,
		},
		{
			"struct2 length",
			struct2{4, []string{"1", "2", "3"}, true},
			false,
		},
		{
			"struct1 pointer length",
			&struct2{4, []string{"1", "2", "3"}, true},
			false,
		},
		{
			"empty string",
			"",
			false,
		},
		{
			"nil",
			nil,
			true,
		},
		{
			"empty array",
			[]string{},
			false,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, types.AnyIsZero(test.obj))
		})
	}
}

func TestAnyValueLen(t *testing.T) {
	t.Parallel()
	type struct1 struct {
		a int
		b []string
	}
	type struct2 struct {
		a int
		b []string
		c bool
	}
	testCases := []struct {
		name        string
		obj         any
		expectedLen int
	}{
		{
			"string length",
			"foobar",
			16,
		},
		{
			"struct1 length",
			struct1{4, []string{"1", "2", "3"}},
			16,
		},
		{
			"struct2 length",
			struct2{4, []string{"1", "2", "3"}, true},
			16,
		},
		{
			"struct1 pointer length",
			&struct2{4, []string{"1", "2", "3"}, true},
			16,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedLen, types.AnyValueLen(test.obj))
		})
	}
}

func TestPrefixEndBytes(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		prefix   []byte
		expected []byte
	}{
		{[]byte{byte(55), byte(255), byte(255), byte(0)}, []byte{byte(55), byte(255), byte(255), byte(1)}},
		{[]byte{byte(55), byte(255), byte(255), byte(15)}, []byte{byte(55), byte(255), byte(255), byte(16)}},
		{[]byte{byte(55), byte(200), byte(255)}, []byte{byte(55), byte(201)}},
		{[]byte{byte(55), byte(255), byte(255)}, []byte{byte(56)}},
		{[]byte{byte(255), byte(255), byte(255)}, nil},
		{[]byte{byte(255)}, nil},
		{nil, nil},
	}

	for _, test := range testCases {
		end := types.PrefixEndBytes(test.prefix)
		assert.DeepEqual(t, test.expected, end)
	}
}

func TestInclusiveEndBytes(t *testing.T) {
	t.Parallel()
	assert.DeepEqual(t, []byte{0x00}, types.InclusiveEndBytes(nil))
	bs := []byte("test")
	assert.DeepEqual(t, append(bs, byte(0x00)), types.InclusiveEndBytes(bs))
}
