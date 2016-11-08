package collector

type MappingRule struct {
	*Rule
	labels []string
}

func NewMappingRule(path []string, labels []string) (*MappingRule, error) {
	result := MappingRule{}
	rule, err := NewRule(path)
	if err != nil {
		return nil, err
	}
	result.Rule = rule
	result.labels = labels
	return &result, nil
}

func (lr *MappingRule) Apply(name []string, labels map[string]string) ([]string, map[string]string) {
	new_name := name[:len(lr.Rule.path)]
	new_labels := labels
	for i := range lr.labels {
		new_labels[lr.labels[i]] = name[len(lr.Rule.path)+i]
	}
	return new_name, new_labels
}
