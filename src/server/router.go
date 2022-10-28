package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func loggingMiddleware(h Handler) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.LogRequest(r)
			next.ServeHTTP(w, r)
		})
	}
}

func basicAuthMiddleware(h Handler, realm string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok {
				// Check if user and password are in query params
				q := r.URL.Query()

				quser, isup := q["username"]
				qpass, ispp := q["password"]

				if !isup || !ispp {
					sendError(w, http.StatusUnauthorized, "Unauthorized")
					return
				}
				user = quser[0]
				pass = qpass[0]
				ok = true
			}

			if !ok || h.CheckBasicAuth(user, pass) {
				sendError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("Not found: %s", r.RequestURI), http.StatusNotFound)
}

func setupRouter(h Handler) *mux.Router {
	r := mux.NewRouter()
	r.Use(loggingMiddleware(h))

	r.Use(basicAuthMiddleware(h, "SIPTV-Optim"))

	r.Path("/{tv}").HandlerFunc(h.FetchTVM3UPlaylist).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return r
}
