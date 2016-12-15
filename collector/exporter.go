package collector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	namespace string
	endpoints     []*EndPoint
	httpworkers   int
	metricfactory *MetricFactory
	metrics       map[string]interface{}
	totascrapes   prometheus.Counter
	mutex         sync.RWMutex
}

func NewExporter(config *Config) *Exporter {
	result := Exporter{}

	result.namespace = config.NameSpace
	result.httpworkers = config.HTTPWorkers

	result.totascrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: result.namespace,
		Name:      "exporter_total_scrapes",
		Help:      "Current total JSON scrapes.",
	})

	result.endpoints = []*EndPoint{}
	for _, endpoint_config := range config.EndPoints {
		endpoint := NewEndPoint(endpoint_config.URI, endpoint_config.Labels, endpoint_config.Interval)
		result.endpoints = append(result.endpoints, endpoint)
	}

	result.metricfactory = NewMetricFactory(config)

	result.metrics = make(map[string]interface{})

	return &result
}

func (e *Exporter) StartRoutines() {
	ch_endpoint := make(chan *EndPoint)
	ch_raw := make(chan *RawMetric)

	for i := 0; i < e.httpworkers; i++ {
		go func() {
			for {
				select {
				case endpoint := <-ch_endpoint:
					content, err_getjson := endpoint.fetchJSONData()
					if err_getjson == nil {
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
					err_status := endpoint.CheckStatus()
					if err_status == nil {
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

	e.totascrapes.Inc()
	ch <- e.totascrapes
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.totascrapes.Desc()
	for _, metric := range e.metrics {
		switch metric.(type) {
		case *prometheus.GaugeVec:
			metric.(*prometheus.GaugeVec).Describe(ch)
		case *prometheus.CounterVec:
			metric.(*prometheus.CounterVec).Describe(ch)
		}
	}
}
