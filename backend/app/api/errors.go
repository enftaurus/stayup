package api

import (
	"github.com/gin-gonic/gin"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func RespondError(c *gin.Context, httpStatus int, code string, message string, field string) {
	c.JSON(httpStatus, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Field:   field,
		},
	})
}
