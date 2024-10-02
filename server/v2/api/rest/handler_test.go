package rest

import (
	"bytes"
	"cosmossdk.io/core/transaction"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDefaultHandlerServeHTTP(t *testing.T) {
	mockAppManager := new(MockAppManager)
	handler := &DefaultHandler[transaction.Tx]{
		appManager: mockAppManager,
	}

	body := []byte(`{"test": "data"}`)
	req, err := http.NewRequest("POST", "/MockMessage", bytes.NewBuffer(body))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	expectedResponse := map[string]string{"result": "success"}
	mockAppManager.On("Query", mock.Anything, int64(0), mock.AnythingOfType("*rest.MockMessage")).Return(expectedResponse, nil)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	mockAppManager.AssertExpectations(t)
}
