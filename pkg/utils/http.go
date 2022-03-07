package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// ValidateHttpRequestMethod valides whether the given http request's method is one of the allowed methods.
// If the method doesn't match, an error message is written back in response.
func ValidateHttpRequestMethod(w http.ResponseWriter, r *http.Request, allowedMethods []string) {
	methodIsValid := false
	for _, method := range allowedMethods {
		if r.Method == method {
			methodIsValid = true
		}
	}
	if !methodIsValid {
		allowedMethodsList := strings.Join(allowedMethods, " or")
		msg := "Method must be: " + allowedMethodsList
		for _, method := range allowedMethods {
			w.Header().Add("Allow", method)
		}
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}
}

// EnableCors writes an Access-Control-Allow-Origin header with the given url to the supplied http response writer, enabling CORS for that url.
func EnableCors(w *http.ResponseWriter, url string) {
	// TODO: needs to check the URL against the list of allowed urls and determine if it matches one of the allowed ones, then send it back
	// this is the recommended way to do it
	// need a different list of allowed URLs for CRUD API calls, Auth service calls, and Health check calls (?)
	(*w).Header().Set("Access-Control-Allow-Origin", url)
}

// SetCorsPreflightResponseHeaders writes the Access-Control-Allow-Methods and Access-Control-Allow-Headers headers to the supplied http response writer.
// This gives the requestor all the information they need to make requests on the particular endpoint you are calling this function from.
func SetCorsPreflightResponseHeaders(w *http.ResponseWriter, allowedMethods []string) {
	for _, method := range allowedMethods {
		(*w).Header().Set("Access-Control-Allow-Methods", method)
	}
	(*w).Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Access-Control-Allow-Credentials, Access-Control-Allow-Origin")
}

// middleware for enabling CORS
func SendCorsPreflightHeaders(allowedUrl string, apiName string, allowedMethods []string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		EnableCors(&w, allowedUrl) // ?? is this the right way to do this
		allowedMethods := []string{http.MethodPost, http.MethodOptions}
		ValidateHttpRequestMethod(w, r, allowedMethods)
		if r.Method == http.MethodOptions {
			SetCorsPreflightResponseHeaders(&w, allowedMethods)
			logrus.Info(fmt.Sprintf(apiName+" API: Sent response to CORS preflight request from %s", r.RemoteAddr))
			return
		}
	})
}

// GetRequestBodyAsString returns the supplied http request's body as a string.
// Returns an empty string and an error if there is a problem reading the http request's body.
func GetRequestBodyAsString(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", errors.New("error reading http request body: " + err.Error())
	}
	return string(body), nil
}

// GetQueryParamAsInt attempts to retrieve the given query parameter by the supplied name, from the supplied http request, and then attempts to convert it to an integer.
// If the parameter is not set, or could not be converted to an integer, -1 and an error is returned. Otherwise, the integer value is returned.
func GetQueryParamAsInt(r *http.Request, paramName string) (int, error) {
	param := r.URL.Query().Get(paramName)
	if param == "" {
		return -1, errors.New("query parameter '" + paramName + "' is not set")
	} else {
		value, err := strconv.Atoi(param)
		if err != nil {
			return -1, errors.New("query parameter '" + paramName + "' could not be converted to an integer: " + param)
		}
		return value, nil
	}
}

// GetQueryParamAsInt attempts to retrieve the given query parameter by the supplied name, from the supplied http request, and then attempts to convert it to an integer.
// If the parameter is not set, or could not be converted to an integer, -1 and an error is returned. Otherwise, the integer value is returned.
func GetQueryParamAsString(r *http.Request, paramName string) (string, error) {
	if r.URL.Query().Has(paramName) {
		return r.URL.Query().Get(paramName), nil
	} else {
		return "", errors.New("query parameter '" + paramName + "' is not set")
	}
}

// PageSeekOptions holds the information required too perform a seek operation against a paginated repository.
type PageSeekOptions struct {
	// The maximum number of records to return.
	RecordLimit int `json:"limit"`
	// The ID to begin the seek operation from.
	StartId int `json:"startid"`
	// The direction to move away from the starting Id.
	Direction string `json:"direction"`
}

const (
	SeekDirectionAfter  = "after"
	SeekDirectionBefore = "before"
	SeekDirectionNone   = "none"
)

// GetPageSeekOptionsByName gets the page seek options for the supplied http request and maximum record limit, using the supplied names for the query parameters.
// Returns the page seek options and the JSON response, which will be an empty struct unless there is an error.
func GetPageSeekOptionsByName(r *http.Request, beforeIdParamName string, afterIdParamName string, limitParamName string, limitMax int) (seekOptions PageSeekOptions, response JsonResponse) {
	var err error
	// Store before id and after id page seek directives
	seekOptions.StartId = -1
	afterIdParamIsSet := r.URL.Query().Has(afterIdParamName)
	if afterIdParamIsSet {
		seekOptions.StartId, err = GetQueryParamAsInt(r, afterIdParamName)
		if err != nil {
			response = JsonResponse{Error: &JsonErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
		}
	}
	beforeIdParamIsSet := r.URL.Query().Has(beforeIdParamName)
	if beforeIdParamIsSet {
		seekOptions.StartId, err = GetQueryParamAsInt(r, beforeIdParamName)
		if err != nil {
			response = JsonResponse{Error: &JsonErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
		}
	}
	if afterIdParamIsSet && beforeIdParamIsSet {
		msg := "Only one of " + afterIdParamName + " and " + beforeIdParamName + " query parameters can be set."
		response = JsonResponse{Error: &JsonErrorResponse{Code: http.StatusBadRequest, Message: msg}}
	} else { // Only one of before id or after id is set, or neither is set
		// Set page seek direction
		if afterIdParamIsSet {
			seekOptions.Direction = SeekDirectionAfter
		} else if beforeIdParamIsSet {
			seekOptions.Direction = SeekDirectionBefore
		} else {
			seekOptions.Direction = SeekDirectionNone
		}

		// Validate page record limit, if set
		seekOptions.RecordLimit = limitMax
		if r.URL.Query().Has(limitParamName) {
			seekOptions.RecordLimit, err = GetQueryParamAsInt(r, limitParamName)
			if err != nil {
				response = JsonResponse{Error: &JsonErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
			} else {
				if seekOptions.RecordLimit > limitMax {
					response = JsonResponse{Error: &JsonErrorResponse{Code: http.StatusBadRequest, Message: limitParamName + " must be less than " + strconv.Itoa(limitMax)}}
				} else if seekOptions.RecordLimit < 1 {
					response = JsonResponse{Error: &JsonErrorResponse{Code: http.StatusBadRequest, Message: limitParamName + " must be greater than 0"}}
				}
			}
		}
	}
	return seekOptions, response
}

// GetPageSeekOptions returns the page seek options for the supplied http request and maximum record limit using standardized names for the query parameters.
func GetPageSeekOptions(r *http.Request, maxLimit int) (PageSeekOptions, JsonResponse) {
	return GetPageSeekOptionsByName(r, "before_id", "after_id", "limit", maxLimit)
}
