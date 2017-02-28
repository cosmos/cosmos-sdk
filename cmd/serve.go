// Copyright © 2017 Ethan Frey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-keys/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the key manager as an http server",
	Long: `Launch an http server with a rest api to manage the
private keys much more in depth than the cli can perform.
In particular, this will allow you to sign transactions with
the private keys in the store.`,
	RunE: serveHTTP,
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntP("port", "p", 8118, "TCP Port for listen for http server")
	serveCmd.Flags().StringP("socket", "s", "", "UNIX socket for more secure http server")
	serveCmd.Flags().StringP("type", "t", "ed25519", "Default key type (ed25519|secp256k1)")
}

func serveHTTP(cmd *cobra.Command, args []string) error {
	var l net.Listener
	var err error
	socket := viper.GetString("socket")
	if socket != "" {
		l, err = createSocket(socket)
		if err != nil {
			return errors.Wrap(err, "Cannot create socket")
		}
	} else {
		port := viper.GetInt("port")
		l, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return errors.Errorf("Cannot listen on port %d", port)
		}
	}

	router := mux.NewRouter()
	ks := server.New(manager, viper.GetString("type"))
	ks.Register(router)

	// only set cors for tcp listener
	var h http.Handler
	if socket == "" {
		allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})
		h = handlers.CORS(allowedHeaders)(router)
	} else {
		h = router
	}

	err = http.Serve(l, h)
	fmt.Printf("Server Killed: %+v\n", err)
	return nil
}

// createSocket deletes existing socket if there, creates a new one,
// starts a server on the socket, and sets permissions to 0600
func createSocket(socket string) (net.Listener, error) {
	err := os.Remove(socket)
	if err != nil && !os.IsNotExist(err) {
		// only fail if it does exist and cannot be deleted
		return nil, err
	}

	l, err := net.Listen("unix", socket)
	if err != nil {
		return nil, err
	}

	mode := os.FileMode(0700) | os.ModeSocket
	err = os.Chmod(socket, mode)
	if err != nil {
		l.Close()
		return nil, err
	}

	return l, nil
}
