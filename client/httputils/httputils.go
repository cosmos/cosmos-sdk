package httputils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewError(ctx *gin.Context, errCode int, err error) {
	errorResponse := HTTPError{
		Api:	"2.0",
		Code:   errCode,
		ErrMsg: err.Error(),
	}
	ctx.JSON(errCode, errorResponse)
}

func Response(ctx *gin.Context, data interface{}) {
	response := HTTPResponse{
		Api:	"2.0",
		Code:   0,
		Result: data,
	}
	ctx.JSON(http.StatusOK, response)
}

type HTTPResponse struct {
	Api 	string 		`json:"rest api" example:"2.0"`
	Code    int    		`json:"code" example:"0"`
	Result 	interface{} `json:"result"`
}

type HTTPError struct {
	Api 	string 		`json:"rest api" example:"2.0"`
	Code    int    		`json:"code" example:"500"`
	ErrMsg 	string 		`json:"error message"`
}