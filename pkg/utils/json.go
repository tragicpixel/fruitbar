package utils

import (
	"github.com/sirupsen/logrus"
	"github.com/tragicpixel/fruitbar/pkg/models"

	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/gddo/httputil/header"
)

const (
	// Maximum size in bytes of a request supplied to the application.
	MAX_CREATE_REQUEST_SIZE_IN_BYTES = 1048576
)

// swagger:response healthCheckResponse
type _ struct {
	body struct {
		Ok bool `json:"ok"`
	}
}

// swagger:response jsonResponse
// Standard response in JSON from the application. (follows Google JSON Format)
type _ struct {
	body JsonResponse
}

// JsonResponse holds a response in JSON format.
type JsonResponse struct {
	// Array of orders returned.
	Data []*models.FruitOrder `json:"data"`
	// Id returned from a newly created object.
	Id string `json:"id"`
	// JSON Web Token returned from completing authentication.
	Token string `json:"token"`
	// Any errors returned by the application.
	Error *JsonErrorResponse `json:"error"`
}

// JsonErrorResponse holds an error response in JSON format. Can contain mulitple errors.
// The top level code/message is used when only one error is contained.
type JsonErrorResponse struct {
	// Top-level HTTP status code returned in the response.
	Code int `json:"code"`
	// Top-level error message.
	Message string `json:"message"`
	// Any additional errors.
	Errors []JsonErrorResponseItem `json:"errors"`
}

// JsonErrorResponseItem holds a single error response in JSON format.
type JsonErrorResponseItem struct {
	// HTTP error code.
	Code string `json:"code"`
	// Error message.
	Message string `json:"message"`
}

// MalformedRequestError holds a response to a malformed request to the application.
type MalformedRequestError struct {
	Status  int
	Message string
}

// Error returns an error string for the error.
func (request *MalformedRequestError) Error() string {
	return request.Message
}

// WriteJSONResponse encodes the specified response object as JSON and then writes it as a response to the supplied response writer, along with the supplied status.
// If there are any errors in encoding or writing, an entry is written to the logs, and an internal server error is written to the page instead.
func WriteJSONResponse(w http.ResponseWriter, status int, response interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		logrus.Error(fmt.Sprintf("Failed to encode json: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("\n"))
		if err != nil {
			logrus.Error(fmt.Sprintf("Failed to write: %s", err.Error()))
		}
	}
}

// DecodeJSONBody decodes a single JSON object into the supplied destination.
// Returns an error if more than one object is included, or there is an error, nil on success.
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, destination interface{}, maxRequestSizeInBytes int) error {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &MalformedRequestError{Status: http.StatusUnsupportedMediaType, Message: msg}
		}
	}

	// Enforce max request size to reduce memory consumption of parsing crazy requests
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxRequestSizeInBytes))

	// Don't ignore extra fields in json -- fail faster, intercept possible erroneous requests
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&destination)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON at position %d", syntaxError.Offset)
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		// Decode() may also return io.ErrUnexpectedEOF for syntax errors in JSON -- open issue in Golang
		// You don't get syntax error position from this though
		// Refactor if https://github.com/golang/go/issues/25956 solved
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field at position %d", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than " + strconv.Itoa(maxRequestSizeInBytes) + " bytes"
			return &MalformedRequestError{Status: http.StatusRequestEntityTooLarge, Message: msg}
		default:
			return err
		}
	}

	// want to rewrite so that multiple json objects of the same type can be decoded
	err = decoder.Decode(&struct{}{}) // if request body only contained a single JSON object this will return io.EOF error
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
	}
	return nil
}
