package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorilla/mux"
)

func TestCommands(t *testing.T) {
	rootCommands := Commands()
	assert.NotNil(t, rootCommands)

	// Commands are registered
	assert.Equal(t, 7, len(rootCommands.Commands()))
}

func TestRegisterRoutes(t *testing.T) {
	fakeRouter := mux.Router{}
	RegisterRoutes(&fakeRouter, false)
}
