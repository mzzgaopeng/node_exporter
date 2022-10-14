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

//go:build !noarp
// +build !noarp

package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	nginxPort = kingpin.Flag("collector.nginx.port", "Ingress nginx stub_status collector").Default("").String()
)

type nginxStatusCollector struct{}

func init() {
	registerCollector("nginx-status", defaultDisabled, NginxCheckCollector)

}

// NewContainerdCollector returns a new Collector.
func NginxCheckCollector() (Collector, error) {
	return &nginxStatusCollector{}, nil

}

func (c *nginxStatusCollector) Update(ch chan<- prometheus.Metric) error {
	nginxports := strings.Split(*nginxPort, ",")
	for _, httpport := range nginxports {
		httpurl := "http://127.0.0.1:" + httpport + "/nginx_status"
		rep, err := http.Get(httpurl)
		if err != nil {
			log.Debugf("Get httpurl faile: %q", httpurl)
			continue
		}

		body, err := ioutil.ReadAll(rep.Body)
		if err != nil {
			log.Debugf("ioutil failed: %q", rep.Body)
		}
		bodystr := strings.Split(string(body), "\n")

		//active metrics
		active := strings.Split(bodystr[0], ": ")
		activemetrics, _ := strconv.Atoi(active[1])

		//server metrics
		servermap := strings.Split(bodystr[2], " ")
		accept, _ := strconv.Atoi(servermap[1])
		handled, _ := strconv.Atoi(servermap[2])
		requests, _ := strconv.Atoi(servermap[3])

		//RWW metrics
		rwwmap := strings.Split(bodystr[3], " ")
		reading, _ := strconv.Atoi(rwwmap[1])
		writing, _ := strconv.Atoi(rwwmap[3])
		waiting, _ := strconv.Atoi(rwwmap[5])

		//active metrics
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "nginx_active"),
				fmt.Sprintf("nginx_active information field %s.", "active"),
				[]string{"httpport"}, nil,
			),
			prometheus.GaugeValue, float64(activemetrics), httpport,
		)
		//server metrics
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "accept"),
				fmt.Sprintf("accept information field %s.", "accept"),
				[]string{"httpport"}, nil,
			),
			prometheus.CounterValue, float64(accept), httpport,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "handled"),
				fmt.Sprintf("handled information field %s.", "handled"),
				[]string{"httpport"}, nil,
			),
			prometheus.CounterValue, float64(handled), httpport,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "requests"),
				fmt.Sprintf("requests information field %s.", "requests"),
				[]string{"httpport"}, nil,
			),
			prometheus.CounterValue, float64(requests), httpport,
		)

		// RWW
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "reading"),
				fmt.Sprintf("requests information field %s.", "reading"),
				[]string{"httpport"}, nil,
			),
			prometheus.GaugeValue, float64(reading), httpport,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "writing"),
				fmt.Sprintf("requests information field %s.", "writing"),
				[]string{"httpport"}, nil,
			),
			prometheus.GaugeValue, float64(writing), httpport,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "waiting"),
				fmt.Sprintf("requests information field %s.", "waiting"),
				[]string{"httpport"}, nil,
			),
			prometheus.GaugeValue, float64(waiting), httpport,
		)

	}
	return nil

}
