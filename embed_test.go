package crypto_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	data "github.com/tendermint/go-wire/data"
)

type PubName struct {
	PubNameInner
}

type PubNameInner interface {
	AssertIsPubNameInner()
	String() string
}

func (p PubName) MarshalJSON() ([]byte, error) {
	return pubNameMapper.ToJSON(p.PubNameInner)
}

func (p *PubName) UnmarshalJSON(data []byte) error {
	parsed, err := pubNameMapper.FromJSON(data)
	if err == nil && parsed != nil {
		p.PubNameInner = parsed.(PubNameInner)
	}
	return err
}

var pubNameMapper = data.NewMapper(PubName{}).
	RegisterImplementation(PubNameFoo{}, "foo", 1).
	RegisterImplementation(PubNameBar{}, "bar", 2)

func (f PubNameFoo) AssertIsPubNameInner() {}
func (f PubNameBar) AssertIsPubNameInner() {}

//----------------------------------------

type PubNameFoo struct {
	Name string
}

func (f PubNameFoo) String() string { return "Foo: " + f.Name }

type PubNameBar struct {
	Age int
}

func (b PubNameBar) String() string { return fmt.Sprintf("Bar #%d", b.Age) }

//----------------------------------------

// TestEncodeDemo tries the various strategies to encode the objects
func TestEncodeDemo(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	cases := []struct {
		in, out  PubNameInner
		expected string
	}{
		{PubName{PubNameFoo{"pub-foo"}}, &PubName{}, "Foo: pub-foo"},
		{PubName{PubNameBar{7}}, &PubName{}, "Bar #7"},
	}

	for i, tc := range cases {

		// Make sure it is proper to start
		require.Equal(tc.expected, tc.in.String())

		// Try to encode as binary
		b, err := data.ToWire(tc.in)
		if assert.Nil(err, "%d: %#v", i, tc.in) {
			err2 := data.FromWire(b, tc.out)
			if assert.Nil(err2) {
				assert.Equal(tc.expected, tc.out.String())
			}
		}

		// Try to encode it as json
		j, err := data.ToJSON(tc.in)
		if assert.Nil(err, "%d: %#v", i, tc.in) {
			err2 := data.FromJSON(j, tc.out)
			if assert.Nil(err2) {
				assert.Equal(tc.expected, tc.out.String())
			}
		}
	}
}
