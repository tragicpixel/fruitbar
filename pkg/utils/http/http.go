package http

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
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
