package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/gogo/protobuf/jsonpb"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
)

const (
	ContentTypeJSON = "application/json"
)

func NewDefaultHandler[T transaction.Tx](appManager *appmanager.AppManager[T]) http.Handler {
	return &DefaultHandler[T]{appManager: appManager}
}

type DefaultHandler[T transaction.Tx] struct {
	appManager *appmanager.AppManager[T]
}

func (h *DefaultHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != ContentTypeJSON {
		contentType = ContentTypeJSON
	}

	requestType := gogoproto.MessageType(path)
	if requestType == nil {
		http.Error(w, "Unknown request type", http.StatusNotFound)
		return
	}

	msg := reflect.New(requestType.Elem()).Interface().(gogoproto.Message)

	err := jsonpb.Unmarshal(r.Body, msg)
	if err != nil {
		http.Error(w, "Error parsing body", http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing body: %v\n", err)
		return
	}

	query, err := h.appManager.Query(r.Context(), 0, msg)
	if err != nil {
		http.Error(w, "Error querying", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(query)
}
