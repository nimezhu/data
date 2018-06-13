package data

import "net/http"

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
func CorsFactory(sites string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Do stuff here
			w.Header().Set("Access-Control-Allow-Origin", sites)
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	}
}
