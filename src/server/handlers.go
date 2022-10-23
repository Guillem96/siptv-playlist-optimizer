package server

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/siptv"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/utils"
	"github.com/gorilla/mux"
)

type Handler struct {
	l    *log.Logger
	conf map[string]siptv.TVConfig
}

func NewHandler(conf map[string]siptv.TVConfig, logger *log.Logger) *Handler {
	return &Handler{
		l:    logger,
		conf: conf,
	}
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.l.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Not found", r.RequestURI)
	http.Error(w, fmt.Sprintf("Not found: %s", r.RequestURI), http.StatusNotFound)
}

func (h *Handler) fetchTVM3UPlaylist(w http.ResponseWriter, r *http.Request) {
	tv, isTvPresent := mux.Vars(r)["tv"]
	if !isTvPresent {
		sendError(w, http.StatusBadRequest, "unexpected error: missing path {tv}.")
		return
	}

	h.l.Printf("%v TV is requesting its channel list.", tv)

	fname := filepath.Join(utils.TempDir(), fmt.Sprintf("%s.m3u", tv))
	exists, err := utils.Exists(fname)
	exists = false
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Error while checking if file exists.")
		return
	}

	if exists {
		h.l.Printf("%v already exists. Forwarding the file for download...", fname)
	} else {
		h.l.Printf("Generating %v m3u playlist...", fname)

		tvConf, isConfPresent := h.conf[tv]
		if !isConfPresent {
			sendError(w, http.StatusNotFound, fmt.Sprintf("%v does not exists.", tv))
			return
		}

		h.l.Printf("Fetching Channels from %+v\n", tvConf.Source)
		channels, err := tvConf.Source.Fetch()
		if err != nil {
			h.l.Printf("Error fetching channels %v", err)
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching channels from %v.", tvConf.Source))
			return
		}

		ocs := siptv.OptimizeChannels(tvConf, channels)
		if err := utils.WriteText(ocs.Marshal(), fname); err != nil {
			h.l.Printf("Error when writing output M3U: %v\n", err)
			sendError(w, http.StatusInternalServerError, "Error while writing generated M3U playlist.")
			return
		}
	}

	http.ServeFile(w, r, fname)
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(message))
}
