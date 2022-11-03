package siptv

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
	"github.com/gorilla/mux"
)

type liveCategories struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	ParentID     int    `json:"parent_id"`
}

type liveStreams struct {
	Num               int    `json:"num"`
	Name              string `json:"name"`
	StreamType        string `json:"stream_type"`
	StreamID          int    `json:"stream_id"`
	StreamIcon        string `json:"stream_icon"`
	EPGCode           string `json:"epg_channel_id"`
	Category          string `json:"category_id"`
	CustomSID         string `json:"custom_sid"`
	TVArchive         int    `json:"tv_archive"`
	TVArchiveDuration int    `json:"tv_archive_duration"`
	DirectSource      string `json:"direct_source"`
}

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
	tvConf, err := h.tvConfFromRequest(r)
	if err != nil {
		utils.SendHTTPError(w, http.StatusBadRequest, err.Error())
		return
	}
	tv := mux.Vars(r)["tv"]
	h.l.Printf("%v TV is requesting its channel list.", tv)

	fname := filepath.Join(utils.TempDir(), fmt.Sprintf("%s.m3u", tv))
	_, err = optimizeAndCacheChannels(tvConf, fname)
	if err != nil {
		h.l.Printf("Error optimizing channels: %s\n", err.Error())
		utils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	http.ServeFile(w, r, fname)
}

func (h *BasicHTTPHandler) PlayerApiHandler(w http.ResponseWriter, r *http.Request) {
	a := r.URL.Query().Get("action")

	tvConf, err := h.tvConfFromRequest(r)
	if err != nil {
		utils.SendHTTPError(w, http.StatusNotFound, err.Error())
		return
	}

	switch a {
	case "get_vod_categories", "get_vod_streams", "get_series_categories", "get_series":
		emptyListHandler(w, r)
	case "get_live_categories":
		h.getLiveCategoriesHandler(tvConf, w, r)
	case "get_live_streams":
		h.getLiveStreamsHandler(tvConf, w, r)
	case "get_simple_data_table":
		h.getEPGInfoHandler(tvConf, w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *BasicHTTPHandler) RedirectToStreamHandler(w http.ResponseWriter, r *http.Request) {
	sid := mux.Vars(r)["streamId"]
	tvConf, err := h.tvConfFromRequest(r)
	if err != nil {
		utils.SendHTTPError(w, http.StatusNotFound, err.Error())
		return
	}

	baseUrl, err := tvConf.Source.BaseStreamUrl()
	if err != nil {
		utils.SendHTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	url := fmt.Sprintf("%s/%s", baseUrl, sid)
	h.l.Println("Redirecting:", url)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func (h *BasicHTTPHandler) getLiveCategoriesHandler(tvConf *TVConfig, w http.ResponseWriter, r *http.Request) {
	var res []liveCategories
	for gn := range tvConf.GroupsFilters {
		res = append(res, liveCategories{gn, gn, 0})
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h *BasicHTTPHandler) getLiveStreamsHandler(tvConf *TVConfig, w http.ResponseWriter, r *http.Request) {
	fname := filepath.Join(utils.TempDir(), fmt.Sprintf("%s.m3u", mux.Vars(r)["tv"]))
	channels, err := optimizeAndCacheChannels(tvConf, fname)
	if err != nil {
		utils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	var res []liveStreams
	for i, c := range channels {
		res = append(res, liveStreams{
			Num:          i,
			Name:         c.ShowName,
			StreamType:   "live",
			StreamID:     c.StreamID,
			StreamIcon:   c.Logo,
			EPGCode:      c.EPGCode,
			Category:     c.Group,
			DirectSource: c.Url,
		})
	}

	if r.URL.Query().Has("category_id") {
		cid := r.URL.Query().Get("category_id")
		var filteredRes []liveStreams
		for _, c := range res {
			if c.Category == cid {
				filteredRes = append(filteredRes, c)
			}
		}
		if err := json.NewEncoder(w).Encode(filteredRes); err != nil {
			utils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		}
	} else {
		if err := json.NewEncoder(w).Encode(res); err != nil {
			utils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		}
	}
}

func (h *BasicHTTPHandler) getEPGInfoHandler(tvConf *TVConfig, w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("stream_id") {
		utils.SendHTTPError(w, http.StatusBadRequest, "Missing stream_id query param")
		return
	}

	sid := r.URL.Query().Get("stream_id")
	url, err := tvConf.Source.EPGUrl(sid)
	if err != nil {
		h.l.Println("Error: cannot form EPG url:", err.Error())
		utils.SendHTTPError(w, http.StatusInternalServerError, "Cannot form EPG url")
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		h.l.Printf("Error forwarding EPG request: %s\n", err.Error())
		utils.SendHTTPError(w, http.StatusInternalServerError, "Error while forwarding the epg request")
		return
	}
	defer resp.Body.Close()

	for k := range resp.Header {
		w.Header().Set(k, resp.Header.Get(k))
	}
	io.Copy(w, resp.Body)
}

func emptyListHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[]"))
	w.Header().Set("Content-Type", "application/json")
}

func optimizeAndCacheChannels(tvConf *TVConfig, cache string) (Playlist, error) {
	exists, err := utils.Exists(cache)
	if err != nil {
		return nil, err
	}

	if exists {
		return Unmarshal(cache)
	}

	channels, err := tvConf.Source.Fetch()
	if err != nil {
		return nil, fmt.Errorf("error fetching channels from %v: %v", tvConf.Source, err.Error())
	}

	ocs := OptimizePlaylist(*tvConf, channels)
	if err := utils.WriteText(ocs.Marshal(), cache); err != nil {
		return nil, errors.New("error while writing generated M3U playlist")
	}
	return ocs, nil
}

func (h *BasicHTTPHandler) tvConfFromRequest(r *http.Request) (*TVConfig, error) {
	tv, isTvPresent := mux.Vars(r)["tv"]
	if !isTvPresent {
		return nil, errors.New("unexpected error: missing path {tv}")
	}

	tvConf, isConfPresent := h.conf[tv]
	if !isConfPresent {
		return nil, fmt.Errorf("%v does not exists", tv)
	}

	return &tvConf, nil
}
