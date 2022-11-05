package siptv

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/configuration"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
)

type PlayListSource interface {
	Fetch() (Playlist, error)
	BaseStreamUrl() (string, error)
	EPGUrl(streamId string) (string, error)
}

type defaultSource struct {
	Username string
	Password string
	Url      string
}

func (s *defaultSource) BaseStreamUrl() (string, error) {
	return fmt.Sprintf("%s/%s/%s", s.Url, s.Username, s.Password), nil
}

func (s *defaultSource) EPGUrl(streamId string) (string, error) {
	url := fmt.Sprintf("%s/player_api.php?username=%s&password=%s&action=get_simple_data_table&stream_id=%s",
		s.Url, s.Username, s.Password, streamId)

	return url, nil
}

type PlayListFileSource struct {
	*defaultSource
	LocalPath string
}

func (s *PlayListFileSource) Fetch() (Playlist, error) {
	return Unmarshal(s.LocalPath)
}

type PlayListUrlSource struct {
	*defaultSource
}

func (s *PlayListUrlSource) Fetch() (Playlist, error) {
	fname := filepath.Join(utils.TempDir(), fmt.Sprintf("%v.m3u", s.Username))
	urlChannels := fmt.Sprintf("%s/get.php?username=%s&password=%s&type=m3u_plus&output=mpegts",
		s.Url, s.Username, s.Password)
	if err := utils.DownloadFile(fname, urlChannels); err != nil {
		return nil, err
	}
	return Unmarshal(fname)
}

type PlayListAPISource struct {
	*defaultSource
}

func (s *PlayListAPISource) Fetch() (cs Playlist, err error) {
	urlChannels := fmt.Sprintf("%s/player_api.php?username=%s&password=%s&action=get_live_streams",
		s.Url, s.Username, s.Password)

	resp, err := http.Get(urlChannels)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	jsonrb, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	streamUrl, err := s.BaseStreamUrl()
	if err != nil {
		return
	}

	err = cs.FromJSON(jsonrb, streamUrl)
	if err != nil {
		return
	}
	return
}

func DigestYAMLSource(source configuration.M3USource) PlayListSource {
	ds := &defaultSource{Username: source.Username, Password: source.Password, Url: source.Url}
	if source.FromFile != "" {
		return &PlayListFileSource{ds, source.FromFile}
	}
	return &PlayListAPISource{ds}
}
