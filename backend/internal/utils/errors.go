// Package utils provides utility functions for the typing master backend
package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleError logs the error and sends a standardized error response
func HandleError(c *gin.Context, status int, err error, message string) {
	log.Printf("Error %d: %s - %v", status, message, err)

	response := ErrorResponse{
		Error:   message,
		Message: err.Error(),
	}

	c.JSON(status, response)
}

// HandleErrorWithCode logs the error and sends a standardized error response with error code
func HandleErrorWithCode(c *gin.Context, status int, err error, message, code string) {
	log.Printf("Error %d (%s): %s - %v", status, code, message, err)

	response := ErrorResponse{
		Error:   message,
		Message: err.Error(),
		Code:    code,
	}

	c.JSON(status, response)
}

// HandleNotFound logs and sends a standardized not found response
func HandleNotFound(c *gin.Context, resource string) {
	message := resource + " not found"
	HandleError(c, http.StatusNotFound, nil, message)
}

// HandleBadRequest logs and sends a standardized bad request response
func HandleBadRequest(c *gin.Context, err error) {
	HandleError(c, http.StatusBadRequest, err, "Invalid request")
}

// HandleInternalError logs and sends a standardized internal server error response
func HandleInternalError(c *gin.Context, err error, operation string) {
	message := "Failed to " + operation
	HandleError(c, http.StatusInternalServerError, err, message)
}

// HandleSuccess sends a standardized success response
func HandleSuccess(c *gin.Context, message string, data interface{}) {
	response := SuccessResponse{
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, response)
}

// HandleCreated sends a standardized created response
func HandleCreated(c *gin.Context, message string, data interface{}) {
	response := SuccessResponse{
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusCreated, response)
}
