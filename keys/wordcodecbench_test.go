package keys

import (
	"testing"

	cmn "github.com/tendermint/tmlibs/common"
)

func warmupCodec(bank string) *WordCodec {
	codec, err := LoadCodec(bank)
	if err != nil {
		panic(err)
	}
	_, err = codec.GetIndex(codec.words[123])
	if err != nil {
		panic(err)
	}
	return codec
}

func BenchmarkCodec(b *testing.B) {
	banks := []string{"english", "spanish", "japanese", "chinese_simplified"}

	for _, bank := range banks {
		b.Run(bank, func(sub *testing.B) {
			codec := warmupCodec(bank)
			sub.ResetTimer()
			benchSuite(sub, codec)
		})
	}
}

func benchSuite(b *testing.B, codec *WordCodec) {
	b.Run("to_words", func(sub *testing.B) {
		benchMakeWords(sub, codec)
	})
	b.Run("to_bytes", func(sub *testing.B) {
		benchParseWords(sub, codec)
	})
}

func benchMakeWords(b *testing.B, codec *WordCodec) {
	numBytes := 32
	data := cmn.RandBytes(numBytes)
	for i := 1; i <= b.N; i++ {
		_, err := codec.BytesToWords(data)
		if err != nil {
			panic(err)
		}
	}
}

func benchParseWords(b *testing.B, codec *WordCodec) {
	// generate a valid test string to parse
	numBytes := 32
	data := cmn.RandBytes(numBytes)
	words, err := codec.BytesToWords(data)
	if err != nil {
		panic(err)
	}

	for i := 1; i <= b.N; i++ {
		_, err := codec.WordsToBytes(words)
		if err != nil {
			panic(err)
		}
	}
}
