// Copyright 2017 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !noarp

package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	defineInfoSubsystem = "define_by_yourself"
)

type defineCollector struct{}

func init() {
	registerCollector("define", defaultEnabled, NewDefineCollector)
}

// NewDefineCollector returns a new Collector.
func NewDefineCollector() (Collector, error) {
	return &defineCollector{}, nil
}

// Update calls update define
func (c *defineCollector) Update(ch chan<- prometheus.Metric) error {

	//Define by yourself

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, defineInfoSubsystem, "define"),
			fmt.Sprintf("define information field %s.", "define"),
			nil, nil,
		),
		prometheus.CounterValue, 1,
	)

	return nil
}
