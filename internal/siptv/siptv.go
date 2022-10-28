package siptv

import (
	"log"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/configuration"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/rules"
)

type TVConfig struct {
	Mac           string
	Source        PlayListSource
	GroupsFilters map[string]rules.Condition
}

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

func OptimizeChannels(conf TVConfig, channels Channels) Channels {
	res := make(Channels, 0)

	for groupName, groupFilter := range conf.GroupsFilters {
		log.Println("Creating group " + groupName + "...")
		filteredChannels := filter(channels, groupFilter)
		for i, cn := range filteredChannels {
			filteredChannels[i] = cn.WithGroupName(groupName)
		}
		log.Printf("%v has %d channels", groupName, len(filteredChannels))

		res = append(res, filteredChannels...)
	}

	return res
}

func filter(channels Channels, cond rules.Condition) (chs Channels) {
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

	groupsToCreate := groups.Definitions
	for _, im := range groups.Imports {
		groupsToCreate[im] = commonGroups[im]
	}

	for groupName, conds := range groupsToCreate {
		res[groupName] = rules.DigestYAMLConditions(conds)
	}

	return res
}
