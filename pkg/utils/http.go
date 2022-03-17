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

// EnableCORS writes an Access-Control-Allow-Origin header with the given url to the supplied http response writer, enabling CORS for that url.
func EnableCORS(w *http.ResponseWriter, url string) {
	// TODO: needs to check the URL against the list of allowed urls and determine if it matches one of the allowed ones, then send it back
	// this is the recommended way to do it
	// need a different list of allowed URLs for CRUD API calls, Auth service calls, and Health check calls (?)
	(*w).Header().Set("Access-Control-Allow-Origin", url)
}

// SetCORSPreflightResponseHeaders writes the Access-Control-Allow-Methods and Access-Control-Allow-Headers headers to the supplied http response writer.
// This gives the requestor all the information they need to make requests on the particular endpoint you are calling this function from.
func SetCORSPreflightResponseHeaders(w *http.ResponseWriter, allowedMethods []string) {
	for _, method := range allowedMethods {
		(*w).Header().Set("Access-Control-Allow-Methods", method)
	}
	(*w).Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Access-Control-Allow-Credentials, Access-Control-Allow-Origin")
}

type CORSOptions struct {
	AllowedUrl     string
	APIName        string
	AllowedMethods []string
}

// middleware for enabling CORS
func SendCORSPreflightHeaders(opts CORSOptions, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(&w, opts.AllowedUrl) // ?? is this the right way to do this
		ValidateHttpRequestMethod(w, r, opts.AllowedMethods)
		if r.Method == http.MethodOptions {
			SetCORSPreflightResponseHeaders(&w, opts.AllowedMethods)
			logrus.Info(fmt.Sprintf("%s API: Sent response to CORS preflight request from %s", opts.APIName, r.RemoteAddr))
			return
		}
		next.ServeHTTP(w, r)
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

// GetQueryParamAsUint returns the value of the given http query parameter from the supplied http request as a uint. (if possible)
func GetQueryParamAsUint(r *http.Request, paramName string) (uint, error) {
	param := r.URL.Query().Get(paramName)
	if param == "" {
		return 0, errors.New("query parameter '" + paramName + "' is not set")
	} else {
		value, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			return 0, errors.New("query parameter '" + paramName + "' could not be converted to an integer: " + param)
		}
		return uint(value), nil
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
	StartId uint `json:"startid"`
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
func GetPageSeekOptionsByName(r *http.Request, beforeIdParam string, afterIdParam string, limitParam string, limitMax int) (opts *PageSeekOptions, err error) {
	opts = &PageSeekOptions{StartId: 0}
	afterIdIsSet := r.URL.Query().Has(afterIdParam)
	beforeIdIsSet := r.URL.Query().Has(beforeIdParam)

	if afterIdIsSet && beforeIdIsSet {
		msg := "Only one of " + afterIdParam + " and " + beforeIdParam + " query parameters can be set."
		return nil, errors.New(msg)
	}

	if afterIdIsSet {
		opts.StartId, err = GetQueryParamAsUint(r, afterIdParam)
		if err != nil {
			return nil, err
		}
		opts.Direction = SeekDirectionAfter
	} else if beforeIdIsSet {
		opts.StartId, err = GetQueryParamAsUint(r, beforeIdParam)
		if err != nil {
			return nil, err
		}
		opts.Direction = SeekDirectionBefore
	} else {
		opts.Direction = SeekDirectionNone
	}

	opts.RecordLimit = limitMax
	if r.URL.Query().Has(limitParam) {
		opts.RecordLimit, err = GetQueryParamAsInt(r, limitParam)
		if err != nil {
			return nil, err
		}
		if opts.RecordLimit > limitMax {
			return nil, fmt.Errorf("%s must be less than %d", limitParam, limitMax)
		} else if opts.RecordLimit < 1 {
			return nil, fmt.Errorf("%s must be greater than 0", limitParam)
		}
	}
	return opts, nil
}

// GetPageSeekOptions returns the page seek options for the supplied http request and maximum record limit using standardized names for the query parameters.
func GetPageSeekOptions(r *http.Request, maxLimit int) (opts *PageSeekOptions, err error) {
	return GetPageSeekOptionsByName(r, "before_id", "after_id", "limit", maxLimit)
}

func WriteJSONErrorResponse(w http.ResponseWriter, status int, errMsg string, logMsg ...string) {
	if logMsg != nil {
		logrus.Error(logMsg)
	} else {
		logrus.Error(errMsg)
	}
	r := JsonResponse{Error: &JsonErrorResponse{Code: status, Message: errMsg}}
	WriteJSONResponse(w, r.Error.Code, r)
}
