package util

import (
	"net/http"
)

// ValidateMethod returns not acceptable for methods other then given.
func ValidateMethod(f http.HandlerFunc, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != string(method) {
			w.WriteHeader(http.StatusNotAcceptable)
		}
		f(w, r)
	}
}
