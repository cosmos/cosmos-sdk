package crypto_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	data "github.com/tendermint/go-data"
)

type Foo struct {
	Name string
}

func (f Foo) Greet() string {
	return "Foo: " + f.Name
}

type Bar struct {
	Age int
}

func (b Bar) Greet() string {
	return fmt.Sprintf("Bar #%d", b.Age)
}

type PubNameInner interface {
	Greet() string
}

type privNameInner interface {
	Greet() string
}

type Greeter interface {
	Greet() string
}

var (
	pubNameMapper, privNameMapper data.Mapper
)

// register both public key types with go-data (and thus go-wire)
func init() {
	pubNameMapper = data.NewMapper(PubName{}).
		RegisterImplementation(Foo{}, "foo", 1).
		RegisterImplementation(Bar{}, "bar", 2)
	privNameMapper = data.NewMapper(PrivName{}).
		RegisterImplementation(Foo{}, "foo", 1).
		RegisterImplementation(Bar{}, "bar", 2)
}

type PubName struct {
	PubNameInner
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

type PrivName struct {
	privNameInner
}

func (p PrivName) MarshalJSON() ([]byte, error) {
	return privNameMapper.ToJSON(p.privNameInner)
}

func (p *PrivName) UnmarshalJSON(data []byte) error {
	parsed, err := privNameMapper.FromJSON(data)
	if err == nil && parsed != nil {
		p.privNameInner = parsed.(privNameInner)
	}
	return err
}

// TestEncodeDemo tries the various strategies to encode the objects
func TestEncodeDemo(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	// assert := assert.New(t)
	// require := require.New(t)

	cases := []struct {
		in, out  Greeter
		expected string
	}{
		{PubName{Foo{"pub-foo"}}, &PubName{}, "Foo: pub-foo"},
		{PubName{Bar{7}}, &PubName{}, "Bar #7"},

		// Note these fail - if you can figure a solution here, I'll buy you a beer :)
		// (ebuchman is right, you must either break the reflection system, or modify go-wire)
		// but such a mod would let us make REALLY sure that no one could construct like this

		// {PrivName{Foo{"priv-foo"}}, &PrivName{}, "Foo: priv-foo"},
		// {PrivName{Bar{9}}, &PrivName{}, "Bar #9"},
	}

	for i, tc := range cases {
		// make sure it is proper to start
		require.Equal(tc.expected, tc.in.Greet())

		// now, try to encode as binary
		b, err := data.ToWire(tc.in)
		if assert.Nil(err, "%d: %#v", i, tc.in) {
			err := data.FromWire(b, tc.out)
			if assert.Nil(err) {
				assert.Equal(tc.expected, tc.out.Greet())
			}
		}

		// try to encode it as json
		j, err := data.ToJSON(tc.in)
		if assert.Nil(err, "%d: %#v", i, tc.in) {
			err := data.FromJSON(j, tc.out)
			if assert.Nil(err) {
				assert.Equal(tc.expected, tc.out.Greet())
			}
		}
	}
}
