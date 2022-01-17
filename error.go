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
