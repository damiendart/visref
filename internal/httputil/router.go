// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package httputil

import (
	"net/http"
	"slices"
)

// The Router type represents an HTTP router that dispatches requests to
// handlers that can be wrapped with middleware functions.
type Router struct {
	middleware []func(http.Handler) http.Handler
	mux        *http.ServeMux
}

// NewRouter returns a new instance of Router.
func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

// Group groups routes within in a Router, allowing middleware to be
// registered for only those routes.
func (router *Router) Group(fn func(Router)) {
	fn(*router)
}

// Handle registers the handler for the given pattern, wrapping the
// handler with middleware functions.
func (router *Router) Handle(pattern string, handler http.Handler) {
	router.mux.Handle(pattern, router.wrapMiddleware(handler))
}

// HandleFunc registers the handler function for the given pattern,
// wrapping the handler with middleware functions.
func (router *Router) HandleFunc(pattern string, fn http.HandlerFunc) {
	router.mux.Handle(pattern, router.wrapMiddleware(fn))
}

// Use registers the given middleware functions. Middleware functions
// are applied to handlers so that they are called in the order they
// were registered.
func (router *Router) Use(m ...func(http.Handler) http.Handler) {
	router.middleware = append(router.middleware, m...)
}

// ServeHTTP makes Router implement the http.Handler interface. It
// implements support for spoofing unsupported HTML form actions (PUT,
// PATCH, and DELETE) with a hidden "_method" input field.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		switch m := r.PostFormValue("_method"); m {
		case http.MethodDelete, http.MethodPatch, http.MethodPut:
			r.Method = m
		}
	}

	router.mux.ServeHTTP(w, r)
}

func (router *Router) wrapMiddleware(h http.Handler) http.Handler {
	for _, mw := range slices.Backward(router.middleware) {
		h = mw(h)
	}

	return h
}
