package collector

type RawMetric struct {
	endpoint *EndPoint
	name     []string
	value    float64
}

func NewRawMetric(endpoint *EndPoint, name []string, value float64) *RawMetric {
	return &RawMetric{
		endpoint: endpoint,
		name:     name,
		value:    value,
	}
}
