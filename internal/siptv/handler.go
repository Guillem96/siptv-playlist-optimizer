package siptv

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
	"github.com/gorilla/mux"
)

type UserCredentials struct {
	Username string
	Password string
}

type BasicHTTPHandler struct {
	l    *log.Logger
	conf map[string]TVConfig
	Auth *UserCredentials
}

func NewBasicHTTPHandler(conf map[string]TVConfig, auth *UserCredentials, logger *log.Logger) *BasicHTTPHandler {
	return &BasicHTTPHandler{
		l:    logger,
		conf: conf,
		Auth: auth,
	}
}

func (h *BasicHTTPHandler) LogRequest(r *http.Request) {
	h.l.Printf("%s: %s", r.Method, r.URL)
}

func (h *BasicHTTPHandler) CheckBasicAuth(reqUsername, reqPassword string) bool {
	if h.Auth == nil {
		return true
	}
	return subtle.ConstantTimeCompare([]byte(h.Auth.Username), []byte(reqUsername)) != 1 ||
		subtle.ConstantTimeCompare([]byte(h.Auth.Password), []byte(reqPassword)) != 1
}

func (h *BasicHTTPHandler) FetchTVM3UPlaylist(w http.ResponseWriter, r *http.Request) {
	tv, isTvPresent := mux.Vars(r)["tv"]
	if !isTvPresent {
		utils.SendHTTPError(w, http.StatusBadRequest, "unexpected error: missing path {tv}.")
		return
	}

	h.l.Printf("%v TV is requesting its channel list.", tv)

	fname := filepath.Join(utils.TempDir(), fmt.Sprintf("%s.m3u", tv))
	exists, err := utils.Exists(fname)
	exists = false
	if err != nil {
		utils.SendHTTPError(w, http.StatusInternalServerError, "Error while checking if file exists.")
		return
	}

	if exists {
		h.l.Printf("%v already exists. Forwarding the file for download...", fname)
	} else {
		h.l.Printf("Generating %v m3u playlist...", fname)

		tvConf, isConfPresent := h.conf[tv]
		if !isConfPresent {
			utils.SendHTTPError(w, http.StatusNotFound, fmt.Sprintf("%v does not exists.", tv))
			return
		}

		h.l.Printf("Fetching Channels from %+v\n", tvConf.Source)
		channels, err := tvConf.Source.Fetch()
		if err != nil {
			h.l.Printf("Error fetching channels %v", err)
			utils.SendHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching channels from %v.", tvConf.Source))
			return
		}

		ocs := OptimizeChannels(tvConf, channels)
		if err := utils.WriteText(ocs.Marshal(), fname); err != nil {
			h.l.Printf("Error when writing output M3U: %v\n", err)
			utils.SendHTTPError(w, http.StatusInternalServerError, "Error while writing generated M3U playlist.")
			return
		}
	}

	http.ServeFile(w, r, fname)
}
