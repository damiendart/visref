package httputil

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRouter_methodSpoofingMethodPrecedence(t *testing.T) {
	mux := NewRouter()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()

	r.Header.Set("X-Http-Method-Override", http.MethodDelete)
	r.PostForm = map[string][]string{"_method": {http.MethodPatch}}

	mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Test", r.Method)
		},
	)

	mux.ServeHTTP(rr, r)

	m := rr.Result().Header.Get("X-Test")

	if m != http.MethodPatch {
		t.Errorf("got %#v, want %#v", m, http.MethodPatch)
	}
}

func TestRouter_methodSpoofingUsingValueFromRequestBody(t *testing.T) {
	for _, tt := range []string{http.MethodDelete, http.MethodPatch, http.MethodPut} {
		mux := NewRouter()
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		rr := httptest.NewRecorder()

		r.PostForm = map[string][]string{"_method": {tt}}

		mux.HandleFunc(
			"/",
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Test", r.Method)
			},
		)

		mux.ServeHTTP(rr, r)

		m := rr.Result().Header.Get("X-Test")

		if m != tt {
			t.Errorf("got %#v, want %#v", m, tt)
		}
	}
}

func TestRouter_methodSpoofingUsingHTTPHeader(t *testing.T) {
	for _, tt := range []string{http.MethodDelete, http.MethodPatch, http.MethodPut} {
		mux := NewRouter()
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		rr := httptest.NewRecorder()

		r.Header.Set("X-Http-Method-Override", tt)

		mux.HandleFunc(
			"/",
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Test", r.Method)
			},
		)

		mux.ServeHTTP(rr, r)

		m := rr.Result().Header.Get("X-Test")

		if m != tt {
			t.Errorf("got %#v, want %#v", m, tt)
		}
	}
}

func TestRouter_routeGrouping(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
	mux := NewRouter()
	newMiddleware := func(value string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("X-Test", value)
					next.ServeHTTP(w, r)
				},
			)
		}
	}

	mux.UseGlobal(newMiddleware("Chicken"))

	mux.Handle("GET /one", handler)

	mux.Use(newMiddleware("Duck"))

	mux.Handle("GET /two", handler)

	mux.Group(
		func(m SubRouter) {
			m.Use(newMiddleware("Goose"))
			m.Handle("GET /three", handler)

			m.Group(
				func(m SubRouter) {
					m.Use(newMiddleware("Pigeon"))
					m.Handle("GET /four", handler)
				},
			)
		},
	)

	mux.Handle("GET /five", handler)

	var tests = []struct {
		requestMethod            string
		requestPath              string
		expectedTestHeaderValues []string
		expectedStatus           int
	}{
		{"GET", "/one", []string{"Chicken"}, http.StatusOK},
		{"GET", "/two", []string{"Chicken", "Duck"}, http.StatusOK},
		{"GET", "/three", []string{"Chicken", "Duck", "Goose"}, http.StatusOK},
		{"GET", "/four", []string{"Chicken", "Duck", "Goose", "Pigeon"}, http.StatusOK},
		{"GET", "/five", []string{"Chicken", "Duck"}, http.StatusOK},
		{"GET", "/404", []string{"Chicken"}, http.StatusNotFound},
		{"POST", "/one", []string{"Chicken"}, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(
			tt.requestPath[1:],
			func(t *testing.T) {
				r, err := http.NewRequest("GET", tt.requestPath, nil)
				if err != nil {
					t.Error(err)
				}

				rr := httptest.NewRecorder()
				mux.ServeHTTP(rr, r)

				v := rr.Result().Header.Values("X-Test")

				if reflect.DeepEqual(v, tt.expectedTestHeaderValues) == false {
					t.Errorf("got %#v, want %#v", v, tt.expectedTestHeaderValues)
				}
			},
		)
	}
}
