package httputils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// NewError create error http response
func NewError(ctx *gin.Context, errCode int, err error) {
	errorResponse := HTTPError{
		API:	"2.0",
		Code:   errCode,
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

// HTTPError is http response with error
type HTTPError struct {
	API 	string 		`json:"rest api" example:"2.0"`
	Code    int    		`json:"code" example:"500"`
	ErrMsg 	string 		`json:"error message"`
}