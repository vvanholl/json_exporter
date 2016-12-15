package collector

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricFactory struct {
	whitelist []*Rule
	blacklist []*Rule
	mapping   []*MappingRule
}

func NewMetricFactory(config *Config) *MetricFactory {
	result := MetricFactory{}

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
	return strings.Replace(strings.Join(name, separator), ".", separator, -1)
}

func (mf *MetricFactory) FilterWhiteList(rawmetric *RawMetric) bool {
	if len(mf.whitelist) != 0 {
		for _, rule := range mf.whitelist {
			if rule.Match(rawmetric.name) {
				return true
			}
		}
		return false
	}
	return true
}

func (mf *MetricFactory) FilterBlackList(rawmetric *RawMetric) bool {
	for _, rule := range mf.blacklist {
		if rule.Match(rawmetric.name) {
			return false
		}
	}
	return true
}

func (mf *MetricFactory) ApplyMapping(rawmetric *RawMetric) ([]string, map[string]string) {
	name := rawmetric.name
	labels := rawmetric.endpoint.labels
	for _, rule := range mf.mapping {
		if rule.Match(rawmetric.name) {
			return rule.Apply(name, labels)
		}
	}
	return name, labels
}

func (mf *MetricFactory) ProcessRawMetric(rawmetric *RawMetric, metrics map[string]interface{}) {
	if mf.FilterWhiteList(rawmetric) && mf.FilterBlackList(rawmetric) {

		name, labels := mf.ApplyMapping(rawmetric)

		labels_k := []string{}
		labels_v := []string{}
		for k, v := range labels {
			labels_k = append(labels_k, k)
			labels_v = append(labels_v, v)
		}

		fmt.Println(labels_k)
		metric_name := mf.GetMetricName(name)

		metric, exists := metrics[metric_name]
		if !exists {
			metric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: "",
				Name:      metric_name,
				Help:      "No Help provided",
			},
				labels_k,
			)
			metrics[metric_name] = metric
		}
		metric.(*prometheus.GaugeVec).WithLabelValues(labels_v...).Set(rawmetric.value)
	}
}
