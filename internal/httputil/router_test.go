package httputil

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRouterGroup(t *testing.T) {
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

	mux.Use(newMiddleware("Chicken"))
	mux.Handle("GET /one", handler)

	mux.Group(
		func(m Router) {
			m.Use(newMiddleware("Duck"))
			m.Handle("GET /two", handler)
		},
	)

	mux.Handle("GET /three", handler)

	var tests = []struct {
		path                 string
		expectedHeaderValues []string
	}{
		{"/one", []string{"Chicken"}},
		{"/two", []string{"Chicken", "Duck"}},
		{"/three", []string{"Chicken"}},
	}

	for _, tt := range tests {
		t.Run(
			tt.path[1:],
			func(t *testing.T) {
				r, err := http.NewRequest("GET", tt.path, nil)
				if err != nil {
					t.Error(err)
				}

				rr := httptest.NewRecorder()
				mux.ServeHTTP(rr, r)

				v := rr.Result().Header.Values("X-Test")

				if reflect.DeepEqual(v, tt.expectedHeaderValues) == false {
					t.Errorf("got %#v, want %#v", tt.expectedHeaderValues, v)
				}
			},
		)
	}
}
