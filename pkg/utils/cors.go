package utils

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

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
