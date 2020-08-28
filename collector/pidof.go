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
	"gopkg.in/alecthomas/kingpin.v2"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	pidofnfo = "shell_pidof_status"
)

var (
	pidNAME = kingpin.Flag("collector.pidof.name", "pidof for pid collector").String()
)

type pidofCollector struct{}

func init() {
	registerCollector("pidof", defaultDisabled, NewpidofCollector)
}

// NewPidofollector returns a new Collector.
func NewpidofCollector() (Collector, error) {
	return &pidofCollector{}, nil
}

// Update calls update pidof
func (c *pidofCollector) Update(ch chan<- prometheus.Metric) error {

	pidofStatus := 0
	pidname := string(*pidNAME)

	//pidof by yourself
	out, err := exec.Command("pidof", pidname).Output()

	if err != nil {
		log.Debugf("Get  %s pid faile: %q", pidname, err)
	}

	if len(out) == 0 {
		pidofStatus = 1
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, pidofnfo, "pidof"),
			fmt.Sprintf("pidof information field %s.", pidname),
			nil, nil,
		),
		prometheus.CounterValue, float64(pidofStatus),
	)

	return nil
}
