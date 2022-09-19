package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	rt6statSubsystem = "rt6stat"
)

type rt6statCollector struct{}

func init() {
	registerCollector("rt6stat", defaultEnabled, NewRt6statCollector)
}

func NewRt6statCollector() (Collector, error) {
	return &rt6statCollector{}, nil
}

func (c *rt6statCollector) Update(ch chan<- prometheus.Metric) error {
	var metricType prometheus.ValueType
	rt6Info, err := c.getRt6stat()
	if err != nil {
		return fmt.Errorf("couldn't get rt6Info: %s", err)
	}
	log.Debugf("Set node_rt6: %#v", rt6Info)

	customLabel := prometheus.Labels{
		"use": fmt.Sprintf("%.0f", rt6Info["rt6_use"]),
		"max": fmt.Sprintf("%.0f", rt6Info["rt6_total"]),
	}
	var usage float64 = 0
	if rt6Info["rt6_total"] > 0 {
		usage = rt6Info["rt6_use"] / rt6Info["rt6_total"]
	}
	metricType = prometheus.GaugeValue
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, rt6statSubsystem, "rt6_usage"),
			fmt.Sprintf("Route6 usage information field %s.", "rt6_usage"),
			nil, customLabel,
		),
		metricType, usage,
	)
	return nil
}
