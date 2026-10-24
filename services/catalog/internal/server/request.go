package server

import (
	"net/http"
	"strconv"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// parsePagination extracts limit and offset query parameters from r, applying safe defaults and bounds.
func parsePagination(r *http.Request) (limit, offset int32) {
	q := r.URL.Query()

	limit = defaultLimit
	if l, err := strconv.ParseInt(q.Get("limit"), 10, 32); err == nil && l > 0 {
		limit = min(int32(l), maxLimit)
	}

	if o, err := strconv.ParseInt(q.Get("offset"), 10, 32); err == nil && o >= 0 {
		offset = int32(o)
	}

	return limit, offset
}
