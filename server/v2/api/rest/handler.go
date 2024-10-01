package rest

import (
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
	"encoding/json"
	"fmt"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"io"
	"net/http"
	"strings"
)

type DefaultHandler[T transaction.Tx] struct {
	appManager *appmanager.AppManager[T]
}

func (h *DefaultHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	requestType := gogoproto.MessageType(r.URL.Path)
	if requestType == nil {
		http.Error(w, "Unknown request type", http.StatusNotFound)
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error parsing body", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Ruta accedida: %s\n", path)
	fmt.Fprintf(w, "Datos recibidos:\n")

	json.NewEncoder(w).Encode(data)
}
