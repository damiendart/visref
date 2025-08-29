package httputil

import (
	"net/http"
)

// ChainableHandler is a chainable [http.Handler] implementation.
type ChainableHandler func(response http.ResponseWriter, r *http.Request) ChainableHandler

// ServeHTTP makes [ChainableHandler] implement the [http.Handler]
// interface and runs the chain until nil is returned.
func (h ChainableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if next := h(w, r); next != nil {
		next.ServeHTTP(w, r)
	}
}
