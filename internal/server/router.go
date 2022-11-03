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
				user = q.Get("username")
				pass = q.Get("password")
				ok = q.Has("username") && q.Has("password")
			}

			if !ok {
				var isUserPresent bool
				var isPasswordPresent bool

				user, isUserPresent = mux.Vars(r)["user"]
				pass, isPasswordPresent = mux.Vars(r)["password"]
				ok = isUserPresent && isPasswordPresent
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
	fmt.Println("NOT FOUND", r.URL)
	http.Error(w, fmt.Sprintf("Not found: %s", r.RequestURI), http.StatusNotFound)
}

func setupRouter(h Handler) *mux.Router {
	r := mux.NewRouter()
	r.Use(buildLoggingMiddleware(h))
	r.Use(buildBasicAuthMiddleware(h, "SIPTV-Optimized"))

	r.Path("/{tv}/player_api.php").HandlerFunc(h.PlayerApiHandler).Methods("GET")
	r.Path("/{tv}/live/{user}/{password}/{streamId}.ts").HandlerFunc(h.RedirectToStreamHandler).Methods("GET")
	r.Path("/{tv}").HandlerFunc(h.FetchTVM3UPlaylist).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return r
}
