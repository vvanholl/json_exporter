package collector

import (
	"regexp"
)

type Rule struct {
	path          []string
	path_compiled []*regexp.Regexp
}

func NewRule(path []string) (*Rule, error) {
	result := Rule{}
	result.path = path
	for _, p := range path {
		compiled, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		result.path_compiled = append(result.path_compiled, compiled)
	}
	return &result, nil
}

func (r *Rule) Match(name []string) bool {
	for i := range r.path_compiled {
		if !r.path_compiled[i].MatchString(name[i]) {
			return false
		}
	}
	return true
}
