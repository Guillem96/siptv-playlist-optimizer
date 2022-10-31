package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
)

func buildLoggingMiddleware(h Handler) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.LogRequest(r)
			next.ServeHTTP(w, r)
		})
	}
}

func buildBasicAuthMiddleware(h Handler, realm string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok {
				// Check if user and password are in query params
				q := r.URL.Query()

				quser, isup := q["username"]
				qpass, ispp := q["password"]

				if !isup || !ispp {
					utils.SendHTTPError(w, http.StatusUnauthorized, "Unauthorized")
					return
				}
				user = quser[0]
				pass = qpass[0]
				ok = true
			}

			if !ok || h.CheckBasicAuth(user, pass) {
				utils.SendHTTPError(w, http.StatusUnauthorized, "Unauthorized")
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
	r.Use(buildLoggingMiddleware(h))
	r.Use(buildBasicAuthMiddleware(h, "SIPTV-Optim"))

	r.Path("/{tv}").HandlerFunc(h.FetchTVM3UPlaylist).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return r
}
