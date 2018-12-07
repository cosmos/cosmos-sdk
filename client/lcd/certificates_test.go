package lcd

import (
	"crypto/ecdsa"
	"crypto/x509"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	host := "127.0.0.1,localhost,::1"
	certBytes, _, err := generateSelfSignedCert(host)
	require.Nil(t, err)
	cert, err := x509.ParseCertificate(certBytes)
	require.Nil(t, err)
	require.Equal(t, 2, len(cert.IPAddresses))
	require.Equal(t, 2, len(cert.DNSNames))
	require.True(t, cert.IsCA)
}

func TestWriteCertAndPrivKey(t *testing.T) {
	expectedPerm := "-rw-------"
	derBytes, priv, err := generateSelfSignedCert("localhost")
	require.Nil(t, err)
	type args struct {
		certBytes []byte
		priv      *ecdsa.PrivateKey
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid certificate", args{derBytes, priv}, false},
		{"garbage", args{[]byte("some garbage"), nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCertFile, gotKeyFile, err := writeCertAndPrivKey(tt.args.certBytes, tt.args.priv)
			defer os.Remove(gotCertFile)
			defer os.Remove(gotKeyFile)
			if tt.wantErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			info, err := os.Stat(gotCertFile)
			require.Nil(t, err)
			require.True(t, info.Mode().IsRegular())
			require.Equal(t, expectedPerm, info.Mode().String())
			info, err = os.Stat(gotKeyFile)
			require.Nil(t, err)
			require.True(t, info.Mode().IsRegular())
			require.Equal(t, expectedPerm, info.Mode().String())
		})
	}
}

func TestFingerprintFromFile(t *testing.T) {
	cert := `-----BEGIN CERTIFICATE-----
MIIBbDCCARGgAwIBAgIQSuFKYv/22v+cxtVgMUrQADAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE4MDkyMDIzNDQyNloXDTE5MDkyMDIzNDQyNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDIo
ujAesRczcPVAWiLhpeV1B7hS/RI2LJaGj3QjyJ8hiUthJTPIamr8m7LuS/U5fS0o
hY297YeTIGo9YkxClICjSTBHMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MA8GA1UdEQQIMAaHBH8AAAEwCgYIKoZI
zj0EAwIDSQAwRgIhAKnwbhX9FrGG1otCVLwhClQ3RaLxnNpCgIGTqSimb34cAiEA
stMN+IqMCKWlZyGqxGIiyksMLMEU3lRqKNQn2EoAZJY=
-----END CERTIFICATE-----`
	wantFingerprint := `SHA256 Fingerprint=0B:ED:9A:AA:A2:D1:7E:B2:53:56:F6:FC:C0:E6:1A:69:70:21:A2:B0:90:FC:AF:BB:EF:AE:2C:78:52:AB:68:40`
	certFile, err := ioutil.TempFile("", "test_cert_")
	require.Nil(t, err)
	_, err = certFile.Write([]byte(cert))
	require.Nil(t, err)
	err = certFile.Close()
	require.Nil(t, err)
	defer os.Remove(certFile.Name())
	fingerprint, err := fingerprintFromFile(certFile.Name())
	require.Nil(t, err)
	require.Equal(t, wantFingerprint, fingerprint)

	// test failure
	emptyFile, err := ioutil.TempFile("", "test_cert_")
	require.Nil(t, err)
	err = emptyFile.Close()
	require.Nil(t, err)
	defer os.Remove(emptyFile.Name())
	_, err = fingerprintFromFile(emptyFile.Name())
	require.NotNil(t, err)
}
