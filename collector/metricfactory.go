package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricFactory struct {
	metrics   map[string]interface{}
	whitelist []*Rule
	blacklist []*Rule
	mapping   []*MappingRule
}

func NewMetricFactory(config *Config) *MetricFactory {
	result := MetricFactory{}

	result.metrics = make(map[string]interface{}, 0)

	for _, rule_config := range config.Rules.WhiteList {
		rule, err := NewRule(rule_config.Path)
		if err == nil {
			result.whitelist = append(result.whitelist, rule)
		}
	}

	for _, rule_config := range config.Rules.BlackList {
		rule, err := NewRule(rule_config.Path)
		if err == nil {
			result.blacklist = append(result.blacklist, rule)
		}
	}

	for _, rule_config := range config.Rules.Mapping {
		rule, err := NewMappingRule(rule_config.Path, rule_config.Labels)
		if err == nil {
			result.mapping = append(result.mapping, rule)
		}
	}

	return &result
}

func (mf *MetricFactory) GetMetricName(name []string) string {
	return strings.Replace(strings.Join(name, separator), ".", "_", -1)
}

func (mf *MetricFactory) FilterWhiteList(rawmetric *RawMetric) bool {
	if len(mf.whitelist) != 0 {
		for _, rule := range mf.whitelist {
			if rule.Match(*rawmetric) {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

func (mf *MetricFactory) FilterBlackList(rawmetric *RawMetric) bool {
	for _, rule := range mf.blacklist {
		if rule.Match(*rawmetric) {
			return false
		}
	}
	return true
}

func (mf *MetricFactory) ApplyMapping(rawmetric *RawMetric) ([]string, map[string]string) {
	name := rawmetric.name
	labels := rawmetric.endpoint.labels

	for _, rule := range mf.mapping {
		if rule.Match(*rawmetric) {
			return rule.Apply(name, labels)
		}
	}

	return name, labels
}

func (mf *MetricFactory) ProcessRawMetric(rawmetric *RawMetric) {
	if mf.FilterWhiteList(rawmetric) && mf.FilterBlackList(rawmetric) {

		name, labels := mf.ApplyMapping(rawmetric)

		label_keys := []string{}
		label_values := []string{}
		for k, v := range labels {
			label_keys = append(label_keys, k)
			label_values = append(label_values, v)
		}

		metric_name := mf.GetMetricName(name)

		metric, exists := mf.metrics[metric_name]
		if !exists {
			metric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      metric_name,
				Help:      "No Help provided",
			},
				label_keys,
			)
			mf.metrics[metric_name] = metric
			prometheus.MustRegister(metric.(*prometheus.GaugeVec))
		}
		metric.(*prometheus.GaugeVec).WithLabelValues(label_values...).Set(rawmetric.value)
	}
}
