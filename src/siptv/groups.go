package siptv

import (
	"log"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/src/rules"
)

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
