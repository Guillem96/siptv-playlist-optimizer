package configuration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/pkg/utils"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	CommonGroups map[string][]Condition         `yaml:"commonGroups,omitempty"`
	Tvs          map[string]OptimizeSIPTVConfig `yaml:"tvs"`
}

type OptimizeSIPTVConfig struct {
	Mac    string               `yaml:"mac"`
	Source M3USource            `yaml:"source"`
	Groups GroupsConfigurations `yaml:"groups"`
}

type M3USource struct {
	Url  string `yaml:"url,omitempty"`
	File string `yaml:"file,omitempty"`
}

type GroupsConfigurations struct {
	Definitions map[string][]Condition `yaml:"definitions,omitempty"`
	Imports     []string               `yaml:"imports,omitempty"`
}

type Condition struct {
	Is         string `yaml:"is,omitempty"`
	Contains   string `yaml:"contains,omitempty"`
	NoContains string `yaml:"noContains,omitempty"`
	StartsWith string `yaml:"startswith,omitempty"`
}

func LoadConfiguration(fname string) Configuration {
	var conf Configuration
	yamlFile, err := ioutil.ReadFile(fname)

	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}

	if err = yaml.Unmarshal(yamlFile, &conf); err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	if err = validate(conf); err != nil {
		log.Fatalf("Validation: %v", err)
	}

	return conf
}

func validate(c Configuration) error {
	for groupName, conds := range c.CommonGroups {
		for i, cond := range conds {
			err := validateCondition(cond, fmt.Sprintf("commonGroups.%v[%d]", groupName, i))
			if err != nil {
				return err
			}
		}
	}

	for tvName, tvConf := range c.Tvs {
		if tvConf.Source.File == "" && tvConf.Source.Url == "" {
			return errors.New(fmt.Sprintf("tvs.%v.source.file and tvs.%v.source.url can be empty", tvName, tvName))
		}

		for tvGroupName, tvGroupConds := range tvConf.Groups.Definitions {
			for i, cond := range tvGroupConds {
				err := validateCondition(
					cond,
					fmt.Sprintf("tvs.%v.groups.definitions.%v[%d]", tvName, tvGroupName, i))
				if err != nil {
					return err
				}
			}
		}

		for i, ign := range tvConf.Groups.Imports {
			err := validateImport(
				ign,
				fmt.Sprintf("tvs.%v.groups.imports[%d]", tvName, i),
				maps.Keys(c.CommonGroups))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateImport(ig, key string, commonGroupsNames []string) error {
	if !utils.Contains(commonGroupsNames, ig) {
		return errors.New(
			"Invalid group import in " + key + ": " + ig +
				" is not present in " + strings.Join(commonGroupsNames, ","))
	}
	return nil
}

func validateCondition(c Condition, key string) error {
	if c.Is != "" && (c.Contains != "" || c.StartsWith != "" || c.NoContains != "") {
		return errors.New("Invalid condition in " + key + ": If `is` is provided" +
			"`startswith`, `contains` and `noContains` must be empty.")
	}

	if c.Is == "" && c.Contains == "" && c.StartsWith == "" {
		return errors.New(
			"Invalid condition in " + key + ": One of `is`, `contains`, `noContains` or `startswith` needs a value.")
	}

	return nil
}