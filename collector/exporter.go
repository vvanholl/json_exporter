package collector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	namespace            string
	num_endpoint_workers int
	endpoints            []*EndPoint
	metricfactory        *MetricFactory
	metrics              map[string]interface{}
	totalscrapes         prometheus.Counter
	mutex                sync.RWMutex
}

func NewExporter(config *Config) (*Exporter, error) {
	var err error

	result := Exporter{
		namespace:            config.NameSpace,
		num_endpoint_workers: config.NumEndpointWorkers,
		endpoints:            []*EndPoint{},
		metrics:              make(map[string]interface{}),
		totalscrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: config.NameSpace,
			Name:      "exporter_totalscrapes",
			Help:      "Current total JSON scrapes.",
		}),
	}

	for _, endpoint_config := range config.EndPoints {
		endpoint := NewEndPoint(endpoint_config.URI, endpoint_config.Labels, endpoint_config.Interval)
		result.endpoints = append(result.endpoints, endpoint)
	}

	result.metricfactory, err = NewMetricFactory(config)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (e *Exporter) StartRoutines() {
	ch_endpoint := make(chan *EndPoint)
	ch_raw := make(chan *RawMetric)

	for i := 0; i < e.num_endpoint_workers; i++ {
		go func() {
			for {
				select {
				case endpoint := <-ch_endpoint:
					content, err := endpoint.fetchJSONData()
					if err == nil {
						endpoint.JSONToRawMetrics([]string{e.namespace}, content, ch_raw)
					}
					endpoint.setStatusWait()
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case rawmetric := <-ch_raw:
				e.metricfactory.ProcessRawMetric(rawmetric, e.metrics)
			}
		}
	}()

	ticker := time.Tick(time.Millisecond * 100)

	go func() {
		for {
			select {
			case <-ticker:
				for _, endpoint := range e.endpoints {
					if endpoint.CheckStatus() == nil {
						endpoint.setStatusQueue()
						ch_endpoint <- endpoint
					}
				}
			}
		}
	}()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, metric := range e.metrics {
		switch metric.(type) {
		case *prometheus.GaugeVec:
			metric.(*prometheus.GaugeVec).Collect(ch)
		case *prometheus.CounterVec:
			metric.(*prometheus.CounterVec).Collect(ch)
		}
	}

	e.totalscrapes.Inc()
	ch <- e.totalscrapes
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.totalscrapes.Desc()

	for _, metric := range e.metrics {
		switch metric.(type) {
		case *prometheus.GaugeVec:
			metric.(*prometheus.GaugeVec).Describe(ch)
		case *prometheus.CounterVec:
			metric.(*prometheus.CounterVec).Describe(ch)
		}
	}
}
