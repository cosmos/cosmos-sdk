package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/cosmos/gogoproto/jsonpb"
	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/appmanager"
)

const (
	ContentTypeJSON   = "application/json"
	MaxBodySize       = 1 << 20 // 1 MB
	BlockHeightHeader = "x-cosmos-block-height"
)

func NewDefaultHandler[T transaction.Tx](stf appmanager.StateTransitionFunction[T], store serverv2.Store, gasLimit uint64) http.Handler {
	return &DefaultHandler[T]{
		stf:      stf,
		store:    store,
		gasLimit: gasLimit,
	}
}

type DefaultHandler[T transaction.Tx] struct {
	stf      appmanager.StateTransitionFunction[T]
	store    serverv2.Store
	gasLimit uint64
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

	var reader store.ReaderMap
	height := w.Header().Get(BlockHeightHeader)
	if height == "" {
		_, reader, err = h.store.StateLatest()
		if err != nil {
			http.Error(w, "Error getting latest state", http.StatusInternalServerError)
			return
		}
	} else {
		ih, err := strconv.ParseUint(height, 10, 64)
		if err != nil {
			http.Error(w, "Error parsing block height", http.StatusBadRequest)
			return
		}
		reader, err = h.store.StateAt(ih)
		if err != nil {
			http.Error(w, "Error getting state at height", http.StatusInternalServerError)
			return
		}
	}

	query, err := h.stf.Query(r.Context(), reader, h.gasLimit, msg)
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
