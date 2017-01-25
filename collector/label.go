package collector

type Label struct {
	key   string
	value string
}

type Labels []Label

func NewLabel(key string, value string) *Label {
	return &Label{
		key:   key,
		value: value,
	}
}

func (slice Labels) Len() int {
	return len(slice)
}

func (slice Labels) Less(i, j int) bool {
	return slice[i].key < slice[j].key
}

func (slice Labels) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice Labels) Keys() []string {
	result := []string{}
	for _, l := range slice {
		result = append(result, l.key)
	}
	return result
}

func (slice Labels) Values() []string {
	result := []string{}
	for _, l := range slice {
		result = append(result, l.value)
	}
	return result
}
