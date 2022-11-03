package siptv

import (
	"fmt"
	"path/filepath"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/configuration"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
)

type PlayListSource interface {
	Fetch() (Playlist, error)
	BaseStreamUrl() (string, error)
	EPGUrl(streamId string) (string, error)
}

type PlayListFileSource struct {
	Username  string
	Password  string
	Url       string
	LocalPath string
}

func (s *PlayListFileSource) Fetch() (Playlist, error) {
	return Unmarshal(s.LocalPath)
}

func (s *PlayListFileSource) BaseStreamUrl() (string, error) {
	return fmt.Sprintf("%s/%s/%s", s.Url, s.Username, s.Password), nil
}

func (s *PlayListFileSource) EPGUrl(streamId string) (string, error) {
	url := fmt.Sprintf("%s/player_api.php?username=%s&password=%s&action=get_simple_data_table&stream_id=%s",
		s.Url, s.Username, s.Password, streamId)
	return url, nil
}

type PlayListUrlSource struct {
	Username string
	Password string
	Url      string
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

func (s *PlayListUrlSource) BaseStreamUrl() (string, error) {
	return fmt.Sprintf("%s/%s/%s", s.Url, s.Username, s.Password), nil
}

func (s *PlayListUrlSource) EPGUrl(streamId string) (string, error) {
	url := fmt.Sprintf("%s/player_api.php?username=%s&password=%s&action=get_simple_data_table&stream_id=%s",
		s.Url, s.Username, s.Password, streamId)

	return url, nil
}

func DigestYAMLSource(source configuration.M3USource) PlayListSource {
	if source.FromFile != "" {
		return &PlayListFileSource{source.Username, source.Password, source.Url, source.FromFile}
	}
	return &PlayListUrlSource{source.Username, source.Password, source.Url}
}
