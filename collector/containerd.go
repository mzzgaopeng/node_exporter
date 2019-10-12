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
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	containerdInfo = "shell_containerd_status"
)

type containerdCollector struct{}

func init() {
	registerCollector("containerd", defaultEnabled, NewContainerdCollector)
}

// NewContainerdCollector returns a new Collector.
func NewContainerdCollector() (Collector, error) {
	return &containerdCollector{}, nil
}

// Update calls update containerd
func (c *containerdCollector) Update(ch chan<- prometheus.Metric) error {

	containerdStatus := 0

	//containerd by yourself
	out, err := exec.Command("pidof", "docker-containerd").Output()
	outContainerd, errContainerd := exec.Command("pidof", "containerd").Output()

	if err != nil {
		log.Debugf("Get docker-containerd faile: %q", err)
	}
	if errContainerd != nil {
		log.Debugf("Get containerd faile: %q", errContainerd)
	}

	if len(out) > 0 || len(outContainerd) > 0 {
		containerdStatus = 1
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, containerdInfo, "containerd"),
			fmt.Sprintf("containerd information field %s.", "containerd"),
			nil, nil,
		),
		prometheus.CounterValue, float64(containerdStatus),
	)

	return nil
}
