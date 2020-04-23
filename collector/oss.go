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
	"github.com/prometheus/common/log"
	"os/exec"
)

const (
	OSSCheckInfo = "check_oss_status"
	OSSIPInfo    = "check_oss_ip"
)

type OSSCheckCollector struct{}

func init() {
	registerCollector("oss", defaultDisabled, NewOSSCheckCollector)

}

// NewContainerdCollector returns a new Collector.
func NewOSSCheckCollector() (Collector, error) {
	return &OSSCheckCollector{}, nil

}

// Update calls update osscheck
func (c *OSSCheckCollector) Update(ch chan<- prometheus.Metric) error {
	ossStatus := 0

	//osscheck by yourself
	out, err := exec.Command("/bin/bash", "-c", "curl -s -o /dev/null -w '%{http_code}' http://oss-cn-luoyang-onlinestor-d01-a.res.online.stor --connect-timeout 1 -m 2").Output()

	if err != nil {
		log.Debugf("Get oss check faile: %q", err)
	}

	if string(out) == "000" {
		ossStatus = 1
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, OSSCheckInfo, "osscheck"),
			fmt.Sprintf("osscheck information field %s.", "osscheck"),
			nil, nil,
		),
		prometheus.CounterValue, float64(ossStatus),
	)

	var ossIP string = "0.0.0.0"
	//osscheck by yourself
	outIP, errIP := exec.Command("/bin/bash", "-c", "ping oss-cn-luoyang-onlinestor-d01-a.res.online.stor -c 1 -W 1|head -1|awk '{print $3}'|tr -d '(|)|\n'").Output()

	if errIP != nil {
		log.Debugf("Get oss IP faile: %q", errIP)
	}

	if errIP == nil {
		ossIP = string(outIP)
	}
	if len(outIP) == 0 {
		ossIP = "0.0.0.0"
	}

	fmt.Println("debug ossIP: ", ossIP)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, OSSIPInfo, "ossip"),
			fmt.Sprintf("ossip information field %s.", "ossip"),
			[]string{"ossIP"}, nil,
		),
		prometheus.CounterValue, float64(1), ossIP,
	)

	return nil
}
