package utils

import (
	"net/http"
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

func EnableCors(w *http.ResponseWriter, url string) {
	// needs to check the URL against the list of allowed urls and determine if it matches one of the allowed ones, then send it back
	// this is the recommended way to do it
	// need a different list of allowed URLs for CRUD API calls, Auth service calls, and Health check calls (?)
	(*w).Header().Set("Access-Control-Allow-Origin", url)
}

func SetCorsPreflightResponseHeaders(w *http.ResponseWriter, allowedMethods []string) {
	for _, method := range allowedMethods {
		(*w).Header().Set("Access-Control-Allow-Methods", method)
	}
	(*w).Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Access-Control-Allow-Credentials, Access-Control-Allow-Origin")
}

func SendCorsPreflightHeaders(allowedMethods []string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//r.URL
		// enable cors here (different function)
		// will not using w as a pointer affect this???
		// or just pass w into those functions???
		for _, method := range allowedMethods {
			w.Header().Set("Access-Control-Allow-Methods", method)
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Access-Control-Allow-Credentials, Access-Control-Allow-Origin")
	})
}
