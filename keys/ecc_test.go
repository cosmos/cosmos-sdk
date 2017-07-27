package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cmn "github.com/tendermint/tmlibs/common"
)

var codecs = []ECC{
	NewIBMCRC16(),
	NewSCSICRC16(),
	NewCCITTCRC16(),
	NewIEEECRC32(),
	NewCastagnoliCRC32(),
	NewKoopmanCRC32(),
	NewISOCRC64(),
	NewECMACRC64(),
}

// TestECCPasses makes sure that the AddECC/CheckECC methods are symetric
func TestECCPasses(t *testing.T) {
	assert := assert.New(t)

	checks := append(codecs, NoECC{})

	for _, check := range checks {
		for i := 0; i < 2000; i++ {
			numBytes := cmn.RandInt()%60 + 1
			data := cmn.RandBytes(numBytes)

			checked := check.AddECC(data)
			res, err := check.CheckECC(checked)
			if assert.Nil(err, "%#v: %+v", check, err) {
				assert.Equal(data, res, "%v", check)
			}
		}
	}
}

// TestECCFails makes sure random data will (usually) fail the checksum
func TestECCFails(t *testing.T) {
	assert := assert.New(t)

	checks := codecs
	attempts := 2000

	for _, check := range checks {
		failed := 0
		for i := 0; i < attempts; i++ {
			numBytes := cmn.RandInt()%60 + 1
			data := cmn.RandBytes(numBytes)
			_, err := check.CheckECC(data)
			if err != nil {
				failed += 1
			}
		}
		// we allow up to 1 falsely accepted checksums, as there are random matches
		assert.InDelta(attempts, failed, 1, "%v", check)
	}
}
