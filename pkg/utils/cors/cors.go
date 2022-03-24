package cors

import (
	"fmt"
	"net/http"

	"github.com/tragicpixel/fruitbar/pkg/utils"
	"github.com/tragicpixel/fruitbar/pkg/utils/log"
)

// Enable writes an Access-Control-Allow-Origin header with the given url to the supplied http response writer, enabling CORS for that url.
func Enable(w *http.ResponseWriter, url string) {
	// TODO: needs to check the URL against the list of allowed urls and determine if it matches one of the allowed ones, then send it back
	// this is the recommended way to do it
	// need a different list of allowed URLs for CRUD API calls, Auth service calls, and Health check calls (?)
	(*w).Header().Set("Access-Control-Allow-Origin", url)
}

// SetPreflightHeaders writes the Access-Control-Allow-Methods and Access-Control-Allow-Headers headers to the supplied http response writer.
// This gives the requestor all the information they need to make requests on the particular endpoint you are calling this function from.
func SetPreflightHeaders(w *http.ResponseWriter, allowedMethods []string) {
	for _, method := range allowedMethods {
		(*w).Header().Set("Access-Control-Allow-Methods", method)
	}
	(*w).Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Access-Control-Allow-Credentials, Access-Control-Allow-Origin")
}

type Options struct {
	// URL that is allowed to make CORS requests
	AllowedURL string // TODO: Change this to accept multiple URLs
	// Name of the calling API
	APIName string
	// Allowed HTTP methods to send back for this API
	AllowedMethods []string
}

// SendPreflightHeaders sends the preflight headers for CORS requests.
func SendPreflightHeaders(opts Options, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Enable(&w, opts.AllowedURL) // ?? is this the right way to do this
		utils.ValidateHttpRequestMethod(w, r, opts.AllowedMethods)
		if r.Method == http.MethodOptions {
			SetPreflightHeaders(&w, opts.AllowedMethods)
			log.Info(fmt.Sprintf("%s API: Sent response to CORS preflight request from %s", opts.APIName, r.RemoteAddr))
			return
		}
		next.ServeHTTP(w, r)
	})
}