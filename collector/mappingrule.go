package collector

type MappingRule struct {
	*Rule
	labels []string
}

func NewMappingRule(path []string, labels []string) (*MappingRule, error) {
	rule, err := NewRule(path)
	if err != nil {
		return nil, err
	}
	return &MappingRule{
		Rule:   rule,
		labels: labels,
	}, nil
}

func (mr *MappingRule) Apply(old_name []string, old_labels Labels) ([]string, Labels) {
	if len(old_name) == len(mr.path)+len(mr.labels) {
		new_name := old_name[:len(mr.Rule.path)]
		new_labels := old_labels
		for i := range mr.labels {
			new_labels = append(new_labels, *NewLabel(mr.labels[i], old_name[len(mr.Rule.path)+i]))
		}
		return new_name, new_labels
	} else {
		return old_name, old_labels
	}
}
