package utils

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// WriteErrorResponse prepares and writes a HTTP error
// given a status code and an error message.
func WriteErrorResponse(w *http.ResponseWriter, status int, msg string) {
	(*w).WriteHeader(status)
	(*w).Write([]byte(msg))
}

// NewError create error http response
func NewError(ctx *gin.Context, errCode int, err error) {
	errorResponse := HTTPError{
		API:  "2.0",
		Code: errCode,
	}
	if err != nil {
		errorResponse.ErrMsg = err.Error()
	}

	ctx.JSON(errCode, errorResponse)
}

// NormalResponse create normal http response
func NormalResponse(ctx *gin.Context, data []byte) {
	ctx.Status(http.StatusOK)
	ctx.Writer.Write(data)
}

// HTTPError wrapper error in http response
type HTTPError struct {
	API    string `json:"rest api"`
	Code   int    `json:"code"`
	ErrMsg string `json:"error message"`
}
