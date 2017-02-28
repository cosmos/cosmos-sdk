/*
package proxy provides http handlers to construct a proxy server
for key management, transaction signing, and query validation.

Please read the README and godoc to see how to
configure the server for your application.
*/
package proxy

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	data "github.com/tendermint/go-data"
	"github.com/tendermint/go-keys/proxy/types"

	"github.com/pkg/errors"
)

func readRequest(r *http.Request, o interface{}) error {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return errors.Wrap(err, "Read Request")
	}
	err = json.Unmarshal(data, o)
	if err != nil {
		return errors.Wrap(err, "Parse")
	}
	return validate(o)
}

// most errors are bad input, so 406... do better....
func writeError(w http.ResponseWriter, err error) {
	// fmt.Printf("Error: %+v\n", err)
	res := types.ErrorResponse{
		Code:  406,
		Error: err.Error(),
	}
	writeCode(w, &res, 406)
}

func writeCode(w http.ResponseWriter, o interface{}, code int) {
	// two space indent to make it easier to read
	data, err := data.ToJSON(o)
	if err != nil {
		writeError(w, err)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(data)
	}
}

func writeSuccess(w http.ResponseWriter, o interface{}) {
	writeCode(w, o, 200)
}
