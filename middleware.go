package main

import (
	"github.com/zenazn/goji/web"
	"net/http"
)

// ServerHeader is a simple piece of middleware that sets the Server: header
func ServerHeader(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", FULLVERSION)
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
