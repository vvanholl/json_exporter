package collector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	HTTPWorkers   int
	endpoints     []*EndPoint
	totalScrapes  prometheus.Counter
	mutex         sync.RWMutex
	metricfactory *MetricFactory
}

func NewExporter(config *Config) *Exporter {
	result := Exporter{}

	result.HTTPWorkers = config.HTTPWorkers

	result.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_total_scrapes",
		Help:      "Current total JSON scrapes.",
	})

	result.endpoints = []*EndPoint{}
	for _, endpoint_config := range config.EndPoints {
		endpoint := NewEndPoint(endpoint_config.URI, endpoint_config.Labels, endpoint_config.Interval)
		result.endpoints = append(result.endpoints, endpoint)
	}

	result.metricfactory = NewMetricFactory(config)

	return &result
}

func (e *Exporter) StartRoutines() {
	ch_endpoint := make(chan *EndPoint)
	ch_raw := make(chan *RawMetric)

	for i := 0; i < e.HTTPWorkers; i++ {
		go func() {
			for {
				select {
				case endpoint := <-ch_endpoint:
					content, err_getjson := endpoint.fetchJSONData()
					if err_getjson == nil {
						endpoint.JSONToRawMetrics([]string{namespace}, content, ch_raw)
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
				e.metricfactory.ProcessRawMetric(rawmetric)
			}
		}
	}()

	ticker := time.Tick(time.Millisecond * 100)

	go func() {
		for {
			select {
			case <-ticker:
				for i := 0; i < len(e.endpoints); i++ {
					endpoint := e.endpoints[i]
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

	e.totalScrapes.Inc()
	ch <- e.totalScrapes

}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.totalScrapes.Desc()
}
