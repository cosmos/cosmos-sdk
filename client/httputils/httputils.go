package httputils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// NewError create error http response
func NewError(ctx *gin.Context, errCode int, err error) {
	errorResponse := httpError{
		API:	"2.0",
		Code:   errCode,
		ErrMsg: err.Error(),
	}
	ctx.JSON(errCode, errorResponse)
}

// NormalResponse create normal http response
func NormalResponse(ctx *gin.Context, data interface{}) {
	response := httpResponse{
		API:	"2.0",
		Code:   0,
		Result: data,
	}
	ctx.JSON(http.StatusOK, response)
}

type httpResponse struct {
	API 	string 		`json:"rest api" example:"2.0"`
	Code    int    		`json:"code" example:"0"`
	Result 	interface{} `json:"result"`
}

type httpError struct {
	API 	string 		`json:"rest api" example:"2.0"`
	Code    int    		`json:"code" example:"500"`
	ErrMsg 	string 		`json:"error message"`
}