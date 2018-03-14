package lcd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

const (
	flagBind             = "bind"
	flagCORS             = "cors"
	flagUnsafeConnection = "unsafe_connection"
)

// ServeCommand will generate a long-running rest server
// (aka Light Client Daemon) that exposes functionality similar
// to the cli, but over rest
func ServeCommand() *cobra.Command {
	// TODO get code from app
	cdc := wire.NewCodec()
	cmd := &cobra.Command{
		Use:   "rest-server",
		Short: "Start LCD (light-client daemon), a local REST server",
		RunE:  startRESTServer(cdc),
	}
	// TODO: handle unix sockets also?
	cmd.Flags().StringP(flagBind, "b", "localhost:1317", "Interface and port that server binds to")
	cmd.Flags().String(flagCORS, "", "Set to domains that can make CORS requests (* for all)")
	cmd.Flags().String(flagUnsafeConnection, "false", "Do not enable HTTPS for the REST server")
	cmd.Flags().StringP(client.FlagChainID, "c", "", "ID of chain we connect to")
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	return cmd
}

func startRESTServer(cdc *wire.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bind := viper.GetString(flagBind)
		unsafeConnection, err := strconv.ParseBool(viper.GetString(flagUnsafeConnection))
		if err != nil {
			return err
		}
		r := initRouter(cdc)

		if unsafeConnection {
			return http.ListenAndServe(bind, r)
		}

		// setup https
		// get path to certificates
		ex, err := os.Executable()
		if err != nil {
			return err
		}
		exPath := filepath.Dir(ex)

		_, err = os.Stat(filepath.Join(exPath, "server.crt"))
		os.IsNotExist(err)
		_, err = os.Stat(filepath.Join(exPath, "server.key"))
		os.IsNotExist(err)

		if err != nil {
			err = generateAndSaveCertificate(exPath)
			if err != nil {
				return err
			}
		}

		return http.ListenAndServeTLS(bind, filepath.Join(exPath, "server.crt"), filepath.Join(exPath, "server.key"), r)
	}
}

func pemBlockForKey(priv *ecdsa.PrivateKey) *pem.Block {

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
		os.Exit(2)
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}

}

func generateAndSaveCertificate(exPath string) error {

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	keyOut, err := os.OpenFile(filepath.Join(exPath, "server.key"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Errorf("failed to open server.key for writing:", err)
	}
	pem.Encode(keyOut, pemBlockForKey(privKey))
	keyOut.Close()

	validFor := time.Duration(365 * 24 * time.Hour)
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	template.IPAddresses = append(template.IPAddresses, net.IP("127.0.0.1"))
	template.DNSNames = append(template.DNSNames, "localhost")
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privKey.PublicKey, privKey)
	if err != nil {
		return errors.Errorf("Failed to create certificate: %s", err)
	}
	certOut, err := os.Create(filepath.Join(exPath, "server.crt"))
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	return nil
}

func initRouter(cdc *wire.Codec) http.Handler {
	r := mux.NewRouter()

	// register routes here
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("alive"))
	})

	return r
}
