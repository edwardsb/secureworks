package httpd

import "net/http"

func HealthCheck(endpoint string) func(http.Handler) http.Handler {
	f := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if (r.Method == "GET" || r.Method == "HEAD") && r.URL.Path == endpoint {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return f
}
