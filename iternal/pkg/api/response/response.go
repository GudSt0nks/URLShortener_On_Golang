package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func ValidationError(errs validator.ValidationErrors) Response {
	var result []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			result = append(result, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			result = append(result, fmt.Sprintf("field %s is not an url", err.Field()))
		default:
			result = append(result, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Error(strings.Join(result, ", "))
}
