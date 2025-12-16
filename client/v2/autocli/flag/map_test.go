package flag

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type mockValue struct {
	val string
	err error
}

func (m *mockValue) Set(s string) error {
	if m.err != nil {
		return m.err
	}
	m.val = s
	return nil
}

func (m *mockValue) String() string {
	return m.val
}

func (m *mockValue) Type() string {
	return "mock"
}

func (m *mockValue) Get(_ protoreflect.Value) (protoreflect.Value, error) {
	if m.err != nil {
		return protoreflect.Value{}, m.err
	}
	return protoreflect.ValueOfString(m.val), nil
}

type mockType struct {
	err error
}

func (m *mockType) NewValue(ctx *context.Context, opts *Builder) Value {
	return &mockValue{err: m.err}
}

func (m *mockType) DefaultValue() string {
	return ""
}

func TestCompositeMapValue_Set(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		resolver   keyValueResolver[int]
		valueType  Type
		expectErr  bool
		expectVals map[int]string
	}{
		{
			name:      "valid input",
			input:     "1=foo,2=bar",
			resolver:  strconv.Atoi,
			valueType: &mockType{},
			expectErr: false,
			expectVals: map[int]string{
				1: "foo",
				2: "bar",
			},
		},
		{
			name:      "invalid format",
			input:     "1foo,2=bar",
			resolver:  strconv.Atoi,
			valueType: &mockType{},
			expectErr: true,
		},
		{
			name:      "key resolver fails",
			input:     "1=foo,invalid=bar",
			resolver:  strconv.Atoi,
			valueType: &mockType{},
			expectErr: true,
		},
		{
			name:      "value parsing fails",
			input:     "1=foo,2=bar",
			resolver:  strconv.Atoi,
			valueType: &mockType{err: errors.New("value error")},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			m := &compositeMapValue[int]{
				keyValueResolver: tc.resolver,
				keyType:          "int",
				valueType:        tc.valueType,
				ctx:              &ctx,
				opts:             &Builder{},
				values:           make(map[int]protoreflect.Value),
			}

			err := m.Set(tc.input)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Convert the protoreflect.Value map to a string map for comparison
				actualVals := make(map[int]string)
				for k, v := range m.values {
					actualVals[k] = v.String()
				}

				assert.Equal(t, tc.expectVals, actualVals)
			}
		})
	}
}
