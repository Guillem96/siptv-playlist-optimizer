package siptv

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var TAGS_REGEX = regexp.MustCompile("([a-zA-Z0-9-]+?)=\"([^\"]+)\"")

type Channel struct {
	Name     string
	EPGCode  string
	Logo     string
	Group    string
	ShowName string
	Url      string
	StreamID int
}

type Playlist []*Channel

func (c *Channel) WithGroupName(gn string) *Channel {
	return &Channel{
		Name:     c.Name,
		Logo:     c.Logo,
		Group:    gn,
		ShowName: c.ShowName,
		Url:      c.Url,
		StreamID: c.StreamID,
	}
}

func (c *Channel) Marshal() string {
	return fmt.Sprintf(
		"#EXTINF:-1 tvg-ID=\"%v\" tvg-name=\"%v\" tvg-logo=\"%v\" group-title=\"%v\",%v\n%v\n",
		c.EPGCode, c.Name, c.Logo, c.Group, c.ShowName, c.Url)
}

func (cs Playlist) Marshal() string {
	res := "#EXTM3U\n"
	for _, c := range cs {
		res += c.Marshal()
	}
	return strings.TrimSuffix(res, "\n")
}

func FromText(metadata, url string) (*Channel, error) {
	splitLine := strings.Split(metadata, ",")
	metadata = splitLine[0]
	metadata = strings.TrimPrefix(metadata, "#EXTINF:-1 ")

	attrMatches := TAGS_REGEX.FindAll([]byte(metadata), -1)
	attrMap := make(map[string]string)
	for _, v := range attrMatches {
		splitAttr := strings.Split(string(v), "=")
		attrMap[string(splitAttr[0])] = strings.Trim(string(splitAttr[1]), "\"")
	}

	splitUrl := strings.Split(url, "/")
	streamId, err := strconv.Atoi(splitUrl[len(splitUrl)-1])
	if err != nil {
		streamId = -1
	}

	return &Channel{
		Name:     attrMap["tvg-name"],
		EPGCode:  attrMap["tvg-ID"],
		Logo:     attrMap["tvg-logo"],
		Group:    attrMap["group-title"],
		ShowName: splitLine[1],
		Url:      url,
		StreamID: streamId,
	}, nil
}

func Unmarshal(m3uFile string) (Playlist, error) {
	readFile, err := os.Open(m3uFile)

	if err != nil {
		return nil, fmt.Errorf("reading m3u file: %v", m3uFile)
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() { // Skip header
		if fileScanner.Text() == "#EXTM3U" {
			break
		}
	}

	var cs Playlist
	for fileScanner.Scan() {
		metadata := fileScanner.Text()
		fileScanner.Scan() // Move to url
		nc, err := FromText(metadata, fileScanner.Text())
		if err != nil {
			return nil, fmt.Errorf("parsing line %v: %v", metadata, err.Error())
		}
		cs = append(cs, nc)
	}
	return cs, nil
}
