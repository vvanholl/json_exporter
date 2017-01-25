package collector

import (
	"regexp"
)

type Rule struct {
	path          []string
	path_compiled []*regexp.Regexp
}

func NewRule(path []string) (*Rule, error) {
	result := Rule{
		path:          path,
		path_compiled: make([]*regexp.Regexp, len(path)),
	}
	for i, p := range path {
		compiled, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		result.path_compiled[i] = compiled
	}
	return &result, nil
}

func (r *Rule) Match(name []string) bool {
	for i, compiled := range r.path_compiled {
		if !compiled.MatchString(name[i]) {
			return false
		}
	}
	return true
}
