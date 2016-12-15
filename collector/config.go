package collector

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ConfigEndPoint struct {
	URI      string            `yaml:"url,omitempty"`
	Interval int               `yaml:"interval,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty"`
}

type ConfigRule struct {
	Path []string `yaml:"path,omitempty"`
}

type ConfigMappingRule struct {
	Path   []string `yaml:"path,omitempty"`
	Labels []string `yaml:"labels,omitempty"`
}

type Config struct {
	HTTPWorkers int        `yaml:"httpworkers,omitempty"`
	NameSpace string        `yaml:"namespace,omitempty"`
	Rules       struct {
		WhiteList []ConfigRule        `yaml:"whitelist,omitempty"`
		BlackList []ConfigRule        `yaml:"blacklist,omitempty"`
		Mapping   []ConfigMappingRule `yaml:"mapping,omitempty"`
	} `yaml:"rules,omitempty"`
	Common    ConfigEndPoint   `yaml:"common,omitempty"`
	EndPoints []ConfigEndPoint `yaml:"endpoints,omitempty"`
}

func NewDefaultConfig() *Config {
	return &Config{
		HTTPWorkers: 1,
		NameSpace: "json",
		Common: ConfigEndPoint{
			Interval: 10,
			Labels:   make(map[string]string, 0),
		},
	}
}

func NewFileConfig(filename string) (*Config, error) {
	data, err_read := ioutil.ReadFile(filename)
	if err_read != nil {
		return nil, err_read
	} else {
		config := NewDefaultConfig()
		err_yaml := yaml.Unmarshal([]byte(data), config)
		if err_yaml == nil {
			err_validate := config.Validate()
			if err_validate == nil {
				return config, nil
			} else {
				return nil, err_validate
			}
		} else {
			return nil, err_yaml
		}
	}
}

func (c *Config) Validate() error {
	for i := range c.EndPoints {
		endpoint := &c.EndPoints[i]
		if endpoint.Interval == 0 {
			endpoint.Interval = c.Common.Interval
		}
		if endpoint.URI == "" {
			return fmt.Errorf("Configuration error, all endpoints must have an url")
		}
		if endpoint.Labels == nil {
			endpoint.Labels = make(map[string]string, 0)
		}
		for k, v := range c.Common.Labels {
			_, present := endpoint.Labels[k]
			if !present {
				endpoint.Labels[k] = v
			}
		}
	}
	for i := range c.Rules.WhiteList {
		rule := &c.Rules.WhiteList[i]
		if len(rule.Path) == 0 {
			return fmt.Errorf("Configuration error, all whitelist rules must have a valid path")
		}
	}
	for i := range c.Rules.BlackList {
		rule := &c.Rules.BlackList[i]
		if len(rule.Path) == 0 {
			return fmt.Errorf("Configuration error, all blacklist rules must have a valid path")
		}
	}
	for i := range c.Rules.Mapping {
		rule := &c.Rules.Mapping[i]
		if len(rule.Path) == 0 {
			return fmt.Errorf("Configuration error, all mapping rules must have a valid path")
		}
		if len(rule.Labels) == 0 {
			return fmt.Errorf("Configuration error, all mapping rules must have labels")
		}
	}
	return nil
}
