package logcollect

import (
	"errors"
	"fmt"
	"regexp"

	"public/libs_go/logs"
)

// Compile multiple log line regexes
func ParseLineRegexes(regexStrs []string) ([]*regexp.Regexp, error) {
	regexes := make([]*regexp.Regexp, 0)
	for _, regexStr := range regexStrs {
		regex, err := ParseLineRegex(regexStr)
		if err != nil {
			return regexes, err
		}
		regexes = append(regexes, regex)
	}
	return regexes, nil
}

// Compile a regex & validate expectations for log line parsing
func ParseLineRegex(regexStr string) (*regexp.Regexp, error) {
	// Regex can't be blank
	if regexStr == "" {
		return nil, errors.New("Must provide a regex for parsing log lines; use `--regex.line_regex` flag.")
	}

	// Compile regex
	lineRegex, err := regexp.Compile(regexStr)
	if err != nil {
		logs.Error("Could not compile line regex")
		return nil, err
	}

	// Require at least one named group
	var numNamedGroups int
	for _, groupName := range lineRegex.SubexpNames() {
		if groupName != "" {
			numNamedGroups++
		}
	}
	if numNamedGroups == 0 {
		logs.Error("No named capture groups")
		return nil, errors.New(fmt.Sprintf("No named capture groups found in regex: '%s'. Must provide at least one named group with line regex. Example: `(?P<name>re)`", regexStr))
	}

	return lineRegex, nil
}

type RegexLineParser struct {
	lineRegexes []*regexp.Regexp
}

// RegexLineParser factory
func NewRegexLineParser(regexStrs []string) (*RegexLineParser, error) {
	lineRegexes, err := ParseLineRegexes(regexStrs)
	if err != nil {
		return nil, err
	}
	return &RegexLineParser{lineRegexes}, nil
}

func (p *RegexLineParser) ParseLine(line string) (map[string]interface{}, error) {
	for _, lineRegex := range p.lineRegexes {
		parsed := make(map[string]interface{})
		match := lineRegex.FindAllStringSubmatch(line, -1)
		if match == nil || len(match) == 0 {
			continue // No matches found, skip to next regex
		}

		// Map capture groups
		var firstMatch []string = match[0] // We only care about the first full lineRegex match
		for i, name := range lineRegex.SubexpNames() {
			if i != 0 && i < len(firstMatch) {
				parsed[name] = firstMatch[i]
			}
		}
		return parsed, nil
	}
	return make(map[string]interface{}), nil
}
