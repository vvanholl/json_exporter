package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricFactory struct {
	whitelist []*Rule
	blacklist []*Rule
	mapping   []*MappingRule
}

func NewMetricFactory(config *Config) (*MetricFactory, error) {
	result := MetricFactory{}

	for _, rule_config := range config.Rules.WhiteList {
		rule, err := NewRule(rule_config.Path)
		if err == nil {
			result.whitelist = append(result.whitelist, rule)
		} else {
			return nil, err
		}
	}

	for _, rule_config := range config.Rules.BlackList {
		rule, err := NewRule(rule_config.Path)
		if err == nil {
			result.blacklist = append(result.blacklist, rule)
		} else {
			return nil, err
		}
	}

	for _, rule_config := range config.Rules.Mapping {
		rule, err := NewMappingRule(rule_config.Path, rule_config.Help, rule_config.Labels)
		if err == nil {
			result.mapping = append(result.mapping, rule)
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (mf *MetricFactory) GetMetricName(name []string) string {
	return strings.Replace(strings.Join(name, "_"), ".", "_", -1)
}

func (mf *MetricFactory) FilterWhiteList(name []string) bool {
	if len(mf.whitelist) != 0 {
		for _, rule := range mf.whitelist {
			if rule.Match(name) {
				return true
			}
		}
		return false
	}
	return true
}

func (mf *MetricFactory) FilterBlackList(name []string) bool {
	for _, rule := range mf.blacklist {
		if rule.Match(name) {
			return false
		}
	}
	return true
}

func (mf *MetricFactory) ApplyMapping(name []string, help string, labels Labels) ([]string, string, Labels) {
	for _, rule := range mf.mapping {
		if rule.Match(name) {
			return rule.Apply(name, help, labels)
		}
	}
	return name, help, labels
}

func (mf *MetricFactory) ProcessRawMetric(rawmetric *RawMetric, metrics map[string]interface{}) {
	name := rawmetric.name
	help := "No help provided"
	labels := rawmetric.endpoint.labels

	if mf.FilterWhiteList(name) && mf.FilterBlackList(name) {
		name, help, labels = mf.ApplyMapping(name, help, labels)
		metric_name := mf.GetMetricName(name)
		metric, exists := metrics[metric_name]
		if !exists {
			metric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: "",
				Name:      metric_name,
				Help:      help,
			},
				labels.Keys(),
			)
			metrics[metric_name] = metric
		}
		metric.(*prometheus.GaugeVec).WithLabelValues(labels.Values()...).Set(rawmetric.value)
	}
}
