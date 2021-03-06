package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/gddo/httputil/header"
	"github.com/tragicpixel/fruitbar/pkg/utils/log"
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
	body Response
}

// TODO: Rewrite this and the token logic to just use the data interface, remove the ID member as well
// Response holds a response in JSON format.
type Response struct {
	// The data payload returned by the application, could be a JSON object or string.
	Data interface{} `json:"data"`
	// Id returned from a newly created object.
	Id string `json:"id"`
	// JSON Web Token returned from completing authentication.
	Token string `json:"token"`
	// Any errors returned by the application.
	Error *ErrorResponse `json:"error"`
}

// ErrorResponse holds an error response in JSON format. Can contain mulitple errors.
// The top level code/message is used when only one error is contained.
type ErrorResponse struct {
	// Top-level HTTP status code returned in the response.
	Code int `json:"code"`
	// Top-level error message.
	Message string `json:"message"`
	// Any additional errors.
	Errors []ErrorResponseItem `json:"errors"`
}

// ErrorResponseItem holds a single error response in JSON format.
type ErrorResponseItem struct {
	// HTTP error code.
	Code string `json:"code"`
	// Error message.
	Message string `json:"message"`
}

// MalformedRequestError holds a response to a malformed request to the application.
type MalformedRequestError struct {
	// HTTP status code to send in response.
	Status int
	// Message to send in the response.
	Message string
}

// Error returns an error string for the error.
func (request *MalformedRequestError) Error() string {
	return request.Message
}

// WriteResponse encodes the specified response object as JSON and then writes it as a response to the supplied response writer, along with the supplied status.
// If there are any errors in encoding or writing, an entry is written to the logs, and an internal server error is written to the page instead.
func WriteResponse(w http.ResponseWriter, status int, response interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(fmt.Sprintf("failed to encode json: %s", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("\n"))
		if err != nil { // there was en error writing to the page
			log.Error(fmt.Sprintf("failed to write: %s", err.Error()))
		}
	}
}

// Decode decodes a single JSON object into the supplied destination.
// Returns an error if more than one object is included, or there is an error, nil on success.
func Decode(w http.ResponseWriter, r *http.Request, destination interface{}, maxRequestSizeInBytes int) error {
	// TODO: this should still be an error if the content-type is not specified, as it is not application/json.
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &MalformedRequestError{Status: http.StatusUnsupportedMediaType, Message: msg}
		}
	}

	// Decode the request body into the destination object
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxRequestSizeInBytes)) // Enforce a maximum request size for security
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Don't ignore extra fields in json: fail faster, intercept possible erroneous requests
	err := decoder.Decode(&destination)

	if err != nil {
		// Return various error messages
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("request body contains badly-formed JSON at position %d", syntaxError.Offset)
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		// Decode() may also return io.ErrUnexpectedEOF for syntax errors in JSON -- open issue in Golang
		// You don't get syntax error position from this though
		case errors.Is(err, io.ErrUnexpectedEOF): // TODO: Refactor if https://github.com/golang/go/issues/25956 solved
			msg := "request body contains badly-formed JSON"
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("request body contains an invalid value for the %q field at position %d", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			msg := fmt.Sprintf("request body contains unknown field %s", fieldName)
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		case errors.Is(err, io.EOF):
			msg := "request body must not be empty"
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		case err.Error() == "http: request body too large":
			msg := "request body must not be larger than " + strconv.Itoa(maxRequestSizeInBytes) + " bytes"
			return &MalformedRequestError{Status: http.StatusRequestEntityTooLarge, Message: msg}
		default:
			return err
		}
	} else { // Json was decoded successfully
		// TODO: want to rewrite so that multiple json objects of the same type can be decoded
		// will need to change some other stuff downstream, for example update returns the updated struct, will need to return multiple in this case
		err = decoder.Decode(&struct{}{}) // if request body only contained a single JSON object this will return io.EOF error
		if err != io.EOF {
			msg := "Request body must only contain a single JSON object"
			return &MalformedRequestError{Status: http.StatusBadRequest, Message: msg}
		} else { // There was only one Json object
			return nil
		}
	}
}

// DecodeJSONBodyAndGetResponse attempts to decode the JSON from the supplied http request's body, into the supplied destination object.
// In the event of an error, a standard JSON response for the application will be returned containing all the relevant information about the error.
// If successful, nil will be returned.
func DecodeAndGetErrorResponse(w http.ResponseWriter, r *http.Request, destination interface{}, maxRequestSizeInBytes int) *Response {
	err := Decode(w, r, destination, maxRequestSizeInBytes)
	if err != nil {
		var request *MalformedRequestError
		if errors.As(err, &request) {
			return &Response{Error: &ErrorResponse{Code: request.Status, Message: request.Message}}
		} else { // The error is not a malformed request error
			msg := "Failed to decode JSON body: " + err.Error()
			return &Response{Error: &ErrorResponse{Code: http.StatusInternalServerError, Message: msg}}
		}
	} else { // The json body was successfully decoded
		return &Response{}
	}
}

func NewResponseWithError(code int, msg string) Response {
	return Response{Error: &ErrorResponse{Code: code, Message: msg}}
}

func WriteErrorResponse(w http.ResponseWriter, status int, errMsg string, logMsg ...string) {
	if logMsg != nil {
		log.Error(logMsg)
	} else {
		log.Error(errMsg)
	}
	r := Response{Error: &ErrorResponse{Code: status, Message: errMsg}}
	WriteResponse(w, r.Error.Code, r)
}
