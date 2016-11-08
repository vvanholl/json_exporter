package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	ENDPOINT_STATUS_IDLE   = 0
	ENDPOINT_STATUS_QUEUED = 1
	ENDPOINT_STATUS_WAIT   = 2
)

type EndPoint struct {
	uri             string
	labels          map[string]string
	interval        int
	status          int
	next_check_time time.Time
}

func NewEndPoint(uri string, labels map[string]string, interval int) *EndPoint {
	return &EndPoint{
		uri:             uri,
		labels:          labels,
		interval:        interval,
		status:          ENDPOINT_STATUS_IDLE,
		next_check_time: time.Now(),
	}
}

func (e *EndPoint) CheckStatus() error {
	switch e.status {
	case ENDPOINT_STATUS_IDLE:
		return nil
	case ENDPOINT_STATUS_QUEUED:
		return fmt.Errorf("Endpoint is queued")
	case ENDPOINT_STATUS_WAIT:
		if time.Now().After(e.next_check_time) {
			e.setStatusIdle()
			return nil
		} else {
			return fmt.Errorf("Waiting for endpoint")
		}
	default:
		return fmt.Errorf("Unknown status")
	}
	return nil
}

func (e *EndPoint) setStatusIdle() {
	e.status = ENDPOINT_STATUS_IDLE
}

func (e *EndPoint) setStatusQueue() {
	e.status = ENDPOINT_STATUS_QUEUED
}

func (e *EndPoint) setStatusWait() {
	e.status = ENDPOINT_STATUS_WAIT
	e.next_check_time = time.Now().Add(time.Second * time.Duration(e.interval))
}

func (e *EndPoint) fetchRawData() ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(e.uri)
	if err != nil {
		return nil, err
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		resp.Body.Close()
		return nil, fmt.Errorf("%s, Received HTTP code %d", e.uri, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return body, nil
}

func (e *EndPoint) fetchJSONData() (interface{}, error) {
	rawdata, err := e.fetchRawData()
	if err != nil {
		return nil, err
	}

	var parsed interface{}
	err = json.Unmarshal(rawdata, &parsed)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

func (e *EndPoint) JSONToRawMetrics(metricpath []string, value interface{}, c chan *RawMetric) {
	switch value.(type) {
	case bool:
		rawmetric := NewRawMetric(e, metricpath, map[bool]float64{true: 1.0, false: 0.0}[value.(bool)])
		c <- rawmetric
	case float64:
		rawmetric := NewRawMetric(e, metricpath, value.(float64))
		c <- rawmetric
	case string:
		f, err := strconv.ParseFloat(value.(string), 64)
		if err == nil {
			rawmetric := NewRawMetric(e, metricpath, f)
			c <- rawmetric
		}
	case []interface{}:
		for i := range value.([]interface{}) {
			e.JSONToRawMetrics(append(metricpath, []string{strconv.Itoa(i)}...), value.([]interface{})[i], c)
		}
	case map[string]interface{}:
		for k, v := range value.(map[string]interface{}) {
			e.JSONToRawMetrics(append(metricpath, []string{k}...), v, c)
		}
	}
}
