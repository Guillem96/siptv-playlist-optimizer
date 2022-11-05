package siptv

import (
	"log"
	"sort"
	"strings"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/configuration"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/rules"
)

type TVConfig struct {
	Mac           string
	Source        PlayListSource
	GroupsFilters map[string]rules.Condition
}

// DigestYAMLConfiguration digests a configuration.Configuration object so it can easily
// be used within the siptv package
func DigestYAMLConfiguration(conf configuration.Configuration) map[string]TVConfig {
	res := make(map[string]TVConfig)
	for tv, tvConf := range conf.Tvs {
		res[tv] = TVConfig{
			Mac:           tvConf.Mac,
			Source:        DigestYAMLSource(tvConf.Source),
			GroupsFilters: digestYAMLGroups(tvConf.Groups, conf.CommonGroups),
		}
	}
	return res
}

// OptimizePlaylist given a configuration and a list of channels, removes the clutter
// and simplifies the M3U playlist
func OptimizePlaylist(conf TVConfig, channels Playlist) Playlist {
	res := make(Playlist, 0)

	for groupName, groupFilter := range conf.GroupsFilters {
		log.Println("Creating group " + groupName + "...")
		filteredPlaylist := filter(channels, groupFilter)
		for i, cn := range filteredPlaylist {
			filteredPlaylist[i] = cn.WithGroupName(groupName)
		}
		log.Printf("%v has %d channels", groupName, len(filteredPlaylist))

		res = append(res, filteredPlaylist...)
	}

	// Make it deterministic
	sort.SliceStable(res, func(i, j int) bool {
		return strings.Compare(res[i].ShowName, res[j].ShowName) == -1
	})

	return res
}

func filter(channels Playlist, cond rules.Condition) (chs Playlist) {
	for _, ch := range channels {
		if cond.Apply(ch.Name) {
			chs = append(chs, ch)
		}
	}
	return
}

func digestYAMLGroups(
	groups configuration.GroupsConfigurations,
	commonGroups map[string][]configuration.Condition) map[string]rules.Condition {

	res := make(map[string]rules.Condition)

	if groups.Definitions == nil {
		groups.Definitions = make(map[string][]configuration.Condition)
	}

	groupsToCreate := groups.Definitions
	for _, im := range groups.Imports {
		groupsToCreate[im] = commonGroups[im]
	}

	for groupName, conditions := range groupsToCreate {
		res[groupName] = rules.DigestYAMLConditions(conditions)
	}

	return res
}
