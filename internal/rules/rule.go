package rules

import (
	"regexp"
	"strings"

	"github.com/Guillem96/optimized-m3u-iptv-list-server/internal/configuration"
)

type Condition interface {
	Apply(value string) bool
}

type IsCondition struct {
	EqualsTo string
}

func (c *IsCondition) Apply(value string) bool {
	return value == c.EqualsTo
}

type StartsWithCondition struct {
	Prefix string
}

func (c *StartsWithCondition) Apply(value string) bool {
	return strings.HasPrefix(value, c.Prefix)
}

type ContainsCondition struct {
	SubString string
}

func (c *ContainsCondition) Apply(value string) bool {
	return strings.Contains(value, c.SubString)
}

type NotCondition struct {
	ToNeg Condition
}

// Apply negates the the evaluation of the given condition
func (c *NotCondition) Apply(value string) bool {
	return !c.ToNeg.Apply(value)
}

type RegexpCondition struct {
	Regexp *regexp.Regexp
}

// Apply tires to match the given regular expression
func (c *RegexpCondition) Apply(value string) bool {
	return c.Regexp.Match([]byte(value))
}

type AllCondition struct {
	Conds []Condition
}

// Apply checks all the given conditions evaluate to true
func (c *AllCondition) Apply(value string) bool {
	for _, cond := range c.Conds {
		if !cond.Apply(value) {
			return false
		}
	}
	return true
}

type AnyCondition struct {
	Conds []Condition
}

// Apply checks if any of the given conditions evaluates to true
func (c *AnyCondition) Apply(value string) bool {
	for _, cond := range c.Conds {
		if cond.Apply(value) {
			return true
		}
	}
	return false
}

// DigestYAMLCondition given a raw condition configuration, from the configuration.Condition
// package, converts it to a fancy object oriented like AllCondition that can be evaluated
// with the `Apply` method
func DigestYAMLCondition(confCond configuration.Condition) Condition {
	var res []Condition

	if confCond.Is != "" {
		return &AllCondition{[]Condition{&IsCondition{confCond.Is}}}
	}

	if confCond.Contains != "" {
		res = append(res, &ContainsCondition{confCond.Contains})
	}

	if confCond.StartsWith != "" {
		res = append(res, &StartsWithCondition{confCond.StartsWith})
	}

	if confCond.NoContains != "" {
		res = append(res, &NotCondition{&ContainsCondition{confCond.NoContains}})
	}

	if confCond.Regexp != "" {
		res = append(res, &RegexpCondition{Regexp: regexp.MustCompile(confCond.Regexp)})
	}

	return &AllCondition{res}
}

// DigestYAMLConditions given multiple conditions, from the configuration.Condition
// package, packs them into a single fancy object oriented like AnyCondition that can be evaluated
// with the `Apply` method
func DigestYAMLConditions(confConds []configuration.Condition) Condition {
	var res []Condition
	for _, yamlCond := range confConds {
		res = append(res, DigestYAMLCondition(yamlCond))
	}
	return &AnyCondition{res}
}
