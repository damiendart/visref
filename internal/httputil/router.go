// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package httputil

import (
	"net/http"
	"slices"
)

// A Router is a wrapper around http.ServeMux that adds support for
// middleware functions. Middleware functions chains can be applied
// globally for all requests, including those handled by http.ServeMux
// (e.g. error handlers), and on registered routes.
type Router struct {
	globalMiddleware []func(http.Handler) http.Handler
	routeMiddleware  []func(http.Handler) http.Handler
	serveMux         *http.ServeMux
}

// A SubRouter is used by the Router to create child route groups.
type SubRouter struct {
	middleware []func(http.Handler) http.Handler
	serveMux   *http.ServeMux
}

// NewRouter returns a new instance of Router.
func NewRouter() *Router {
	return &Router{serveMux: http.NewServeMux()}
}

// Group creates a route group, inheriting route middleware chains
// from its parent. It can be extended with additional route middleware.
func (r *Router) Group(fn func(SubRouter)) {
	fn(SubRouter{middleware: slices.Clone(r.routeMiddleware), serveMux: r.serveMux})
}

// Handle registers the handler for the given pattern and wraps the
// handler with the current router middleware chain.
func (r *Router) Handle(pattern string, handler http.Handler) {
	for _, mw := range slices.Backward(r.routeMiddleware) {
		handler = mw(handler)
	}

	r.serveMux.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern and
// wraps the handler with the current router middleware chain.
func (r *Router) HandleFunc(pattern string, fn http.HandlerFunc) {
	r.Handle(pattern, fn)
}

// Use adds the given functions to the route middleware chain.
func (r *Router) Use(m ...func(http.Handler) http.Handler) {
	r.routeMiddleware = append(r.routeMiddleware, m...)
}

// UseGlobal adds the given functions to the global middleware chain.
func (r *Router) UseGlobal(m ...func(http.Handler) http.Handler) {
	r.globalMiddleware = append(r.globalMiddleware, m...)
}

// ServeHTTP makes Router implement the http.Handler interface. It
// implements support for spoofing unsupported HTML form actions (PUT,
// PATCH, and DELETE) with a hidden "_method" input field.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		switch m := req.PostFormValue("_method"); m {
		case http.MethodDelete, http.MethodPatch, http.MethodPut:
			req.Method = m
		}
	}

	var h http.Handler = r.serveMux

	for _, mw := range slices.Backward(r.globalMiddleware) {
		h = mw(h)
	}

	h.ServeHTTP(w, req)
}

// Group creates a route group, inheriting route middleware chains
// from its parent. It can be extended with additional route middleware.
func (s *SubRouter) Group(fn func(router SubRouter)) {
	fn(SubRouter{middleware: slices.Clone(s.middleware), serveMux: s.serveMux})
}

// Handle registers the handler for the given pattern and wraps the
// handler with the current router middleware chain.
func (s *SubRouter) Handle(pattern string, handler http.Handler) {
	for _, mw := range slices.Backward(s.middleware) {
		handler = mw(handler)
	}

	s.serveMux.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern and
// wraps the handler with the current router middleware chain.
func (s *SubRouter) HandleFunc(pattern string, fn http.HandlerFunc) {
	s.Handle(pattern, fn)
}

// Use adds the given functions to the route middleware chain.
func (s *SubRouter) Use(m ...func(http.Handler) http.Handler) {
	s.middleware = append(s.middleware, m...)
}
