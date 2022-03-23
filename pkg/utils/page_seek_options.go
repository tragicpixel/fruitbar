package utils

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tragicpixel/fruitbar/pkg/repository"
)

// GetPageSeekOptionsByName gets the page seek options for the supplied http request and maximum record limit, using the supplied names for the query parameters.
// Returns the page seek options and the JSON response, which will be an empty struct unless there is an error.
func GetPageSeekOptionsByName(r *http.Request, beforeIdParam string, afterIdParam string, limitParam string, limitMax int) (opts *repository.PageSeekOptions, err error) {
	opts = &repository.PageSeekOptions{StartId: 0}
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
		opts.Direction = repository.SeekDirectionAfter
	} else if beforeIdIsSet {
		opts.StartId, err = GetQueryParamAsUint(r, beforeIdParam)
		if err != nil {
			return nil, err
		}
		opts.Direction = repository.SeekDirectionBefore
	} else {
		opts.Direction = repository.SeekDirectionNone
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
func GetPageSeekOptions(r *http.Request, maxLimit int) (opts *repository.PageSeekOptions, err error) {
	return GetPageSeekOptionsByName(r, "before_id", "after_id", "limit", maxLimit)
}
