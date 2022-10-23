package siptv

import (
	"fmt"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/configuration"
	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/rules"
)

type TVConfig struct {
	Mac           string
	Source        PlayListSource
	GroupsFilters map[string]rules.Condition
}

func DigestYAMLConfiguration(conf configuration.Configuration) map[string]TVConfig {
	res := make(map[string]TVConfig)
	for tv, tvConf := range conf.Tvs {
		fmt.Println("-----" + tv)
		res[tv] = TVConfig{
			Mac:           tvConf.Mac,
			Source:        DigestYAMLSource(tvConf.Source),
			GroupsFilters: digestYAMLGroups(tvConf.Groups, conf.CommonGroups),
		}
	}
	return res
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
