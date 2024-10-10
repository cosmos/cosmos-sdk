package rest

import (
	"encoding/json"
	"fmt"
	"io"
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
	MaxBodySize     = 1 << 20 // 1 MB
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

	msg, ok := reflect.New(requestType.Elem()).Interface().(gogoproto.Message)
	if !ok {
		http.Error(w, "Failed to create message instance", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()
	limitedReader := io.LimitReader(r.Body, MaxBodySize)
	err := jsonpb.Unmarshal(limitedReader, msg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing body: %v", err), http.StatusBadRequest)
		return
	}

	query, err := h.appManager.Query(r.Context(), 0, msg)
	if err != nil {
		http.Error(w, "Error querying", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", ContentTypeJSON)
	if err := json.NewEncoder(w).Encode(query); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}
