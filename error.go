package http

import (
	"net/http"
	"strconv"
)

type Error struct {
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	System  string `json:"system,omitempty"`
}

func (e Error) Error() string {
	return "Status: " + strconv.Itoa(e.Status) + " Message: " + e.Message + " Code: " + e.Code + " System: " + e.System
}

func NewError(status int, message string, system string, code int) *Error {
	c := system + "." + strconv.Itoa(status) + strconv.Itoa(code)
	return &Error{
		System:  system,
		Status:  status,
		Message: message,
		Code:    c,
	}
}

func BadRequest(code int, message string, system string) Response {
	err := NewError(http.StatusBadRequest, message, system, code)
	return NewResponse(http.StatusBadRequest, err, nil)
}

func NotFound(code int, message string, system string) Response {
	err := NewError(http.StatusNotFound, message, system, code)
	return NewResponse(http.StatusNotFound, err, nil)
}

func InternalServer(code int, message string, system string) Response {
	err := NewError(http.StatusInternalServerError, message, system, code)
	return NewResponse(http.StatusInternalServerError, err, nil)
}

func Unauthorized(code int, message string, system string) Response {
	err := NewError(http.StatusUnauthorized, message, system, code)
	return NewResponse(http.StatusUnauthorized, err, nil)
}