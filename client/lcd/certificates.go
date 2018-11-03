package lcd

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

// default: 30 days
const defaultValidFor = 30 * 24 * time.Hour

func generateSelfSignedCert(host string) (certBytes []byte, priv *ecdsa.PrivateKey, err error) {
	priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	notBefore := time.Now()
	notAfter := notBefore.Add(defaultValidFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		err = fmt.Errorf("failed to generate serial number: %s", err)
		return
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Gaia Lite"},
		},
		DNSNames:              []string{"localhost"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	certBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		err = fmt.Errorf("couldn't create certificate: %s", err)
		return
	}
	return
}

func writeCertAndPrivKey(certBytes []byte, priv *ecdsa.PrivateKey) (certFile string, keyFile string, err error) {
	if priv == nil {
		err = errors.New("private key is nil")
		return
	}
	certFile, err = writeCertificateFile(certBytes)
	if err != nil {
		return
	}
	keyFile, err = writeKeyFile(priv)
	return
}

func writeCertificateFile(certBytes []byte) (filename string, err error) {
	f, err := ioutil.TempFile("", "cert_")
	if err != nil {
		return
	}
	defer f.Close()
	filename = f.Name()
	if err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return filename, fmt.Errorf("failed to write data to %s: %s", filename, err)
	}
	return
}

func writeKeyFile(priv *ecdsa.PrivateKey) (filename string, err error) {
	f, err := ioutil.TempFile("", "key_")
	if err != nil {
		return
	}
	defer f.Close()
	filename = f.Name()
	block, err := pemBlockForKey(priv)
	if err != nil {
		return
	}
	if err := pem.Encode(f, block); err != nil {
		return filename, fmt.Errorf("failed to write data to %s: %s", filename, err)
	}
	return
}

func pemBlockForKey(priv *ecdsa.PrivateKey) (*pem.Block, error) {
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal ECDSA private key: %v", err)
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil

}

func genCertKeyFilesAndReturnFingerprint(sslHosts string) (certFile, keyFile string, fingerprint string, err error) {
	certBytes, priv, err := generateSelfSignedCert(sslHosts)
	if err != nil {
		return
	}
	certFile, keyFile, err = writeCertAndPrivKey(certBytes, priv)
	cleanupFunc := func() {
		os.Remove(certFile)
		os.Remove(keyFile)
	}
	// Either of the files could have been written already,
	// thus clean up regardless of the error.
	if err != nil {
		defer cleanupFunc()
		return
	}
	fingerprint, err = fingerprintForCertificate(certBytes)
	if err != nil {
		defer cleanupFunc()
		return
	}
	return
}

func fingerprintForCertificate(certBytes []byte) (string, error) {
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(cert.Raw)
	fingerprintBytes := h.Sum(nil)
	var buf bytes.Buffer
	for i, b := range fingerprintBytes {
		if i > 0 {
			fmt.Fprintf(&buf, ":")
		}
		fmt.Fprintf(&buf, "%02X", b)
	}
	return fmt.Sprintf("SHA256 Fingerprint=%s", buf.String()), nil
}

func fingerprintFromFile(certFile string) (string, error) {
	f, err := os.Open(certFile)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return "", fmt.Errorf("couldn't find PEM data in %s", certFile)
	}
	return fingerprintForCertificate(block.Bytes)
}
