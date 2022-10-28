package siptv

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/configuration"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
)

type PlayListSource interface {
	Fetch() (Channels, error)
}

type PlayListFileSource struct {
	LocalPath string
}

func (s *PlayListFileSource) Fetch() (Channels, error) {
	return Unmarshal(s.LocalPath)
}

type PlayListUrlSource struct {
	Url string
}

func (s *PlayListUrlSource) Fetch() (Channels, error) {
	parsedUrl, err := url.Parse(s.Url)
	if err != nil {
		return nil, fmt.Errorf("Error Parsing download url %v", s.Url)
	}

	fname := filepath.Join(
		utils.TempDir(),
		fmt.Sprintf(
			"%v%v.m3u",
			strings.Replace(parsedUrl.RawPath, string(os.PathSeparator), "_", -1),
			parsedUrl.RawQuery,
		),
	)
	utils.DownloadFile(fname, s.Url)
	return Unmarshal(fname)
}

func DigestYAMLSource(source configuration.M3USource) PlayListSource {
	if source.File != "" {
		return &PlayListFileSource{source.File}
	}
	return &PlayListUrlSource{source.Url}
}
