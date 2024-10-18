package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/jsonpb"
	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
)

const (
	ContentTypeJSON = "application/json"
	MaxBodySize     = 1 << 20 // 1 MB
)

func NewDefaultHandler[T transaction.Tx](appManager appmanager.AppManager[T]) http.Handler {
	return &DefaultHandler[T]{appManager: appManager}
}

type DefaultHandler[T transaction.Tx] struct {
	appManager appmanager.AppManager[T]
}

func (h *DefaultHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.validateMethodIsPOST(r); err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	if err := h.validateContentTypeIsJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	msg, err := h.createMessage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

// validateMethodIsPOST validates that the request method is POST.
func (h *DefaultHandler[T]) validateMethodIsPOST(r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed")
	}
	return nil
}

// validateContentTypeIsJSON validates that the request content type is JSON.
func (h *DefaultHandler[T]) validateContentTypeIsJSON(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != ContentTypeJSON {
		return fmt.Errorf("unsupported content type, expected %s", ContentTypeJSON)
	}

	return nil
}

// createMessage creates the message by unmarshalling the request body.
func (h *DefaultHandler[T]) createMessage(r *http.Request) (gogoproto.Message, error) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	requestType := gogoproto.MessageType(path)
	if requestType == nil {
		return nil, fmt.Errorf("unknown request type")
	}

	msg, ok := reflect.New(requestType.Elem()).Interface().(gogoproto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to create message instance")
	}

	defer r.Body.Close()
	limitedReader := io.LimitReader(r.Body, MaxBodySize)
	err := jsonpb.Unmarshal(limitedReader, msg)
	if err != nil {
		return nil, fmt.Errorf("error parsing body: %w", err)
	}

	return msg, nil
}
