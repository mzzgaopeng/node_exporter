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
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type nginxStatusCollector struct{}

var dockerclient client.APIClient

func createcli() {
	var err error
	dockerclient, err = client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}
}

func init() {
	registerCollector("nginx-status", defaultDisabled, NginxCheckCollector)
}

// NewContainerdCollector returns a new Collector.
func NginxCheckCollector() (Collector, error) {
	return &nginxStatusCollector{}, nil

}

func (c *nginxStatusCollector) Update(ch chan<- prometheus.Metric) error {
	createcli()
	containers, err := dockerclient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	nginxmap := make(map[string]string)

	for _, container := range containers {
		log.Infof("containersid  is %s", container.Names)
		for _, containername := range container.Names {
			if strings.Contains(containername, "ingress-controller") && !strings.HasPrefix(containername, "/k8s_POD") {
				log.Debugf("container ingress containerid is %s", container.ID)
				log.Debugf(container.Command)
				split := strings.Fields(container.Command)
				podname := container.Labels["io.kubernetes.pod.name"]
				nginxmap[podname] = strings.TrimLeft(split[3], "--http-port=")
			}
		}
	}
	//nginxports := strings.Split(*nginxPort, ",")
	for containername, httpport := range nginxmap {
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
				[]string{"podname", "httpport"}, nil,
			),
			prometheus.GaugeValue, float64(activemetrics), containername, httpport,
		)
		//server metrics
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "accept"),
				fmt.Sprintf("accept information field %s.", "accept"),
				[]string{"podname", "httpport"}, nil,
			),
			prometheus.CounterValue, float64(accept), containername, httpport,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "handled"),
				fmt.Sprintf("handled information field %s.", "handled"),
				[]string{"podname", "httpport"}, nil,
			),
			prometheus.CounterValue, float64(handled), containername, httpport,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "requests"),
				fmt.Sprintf("requests information field %s.", "requests"),
				[]string{"podname", "httpport"}, nil,
			),
			prometheus.CounterValue, float64(requests), containername, httpport,
		)

		// RWW
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "reading"),
				fmt.Sprintf("requests information field %s.", "reading"),
				[]string{"podname", "httpport"}, nil,
			),
			prometheus.GaugeValue, float64(reading), containername, httpport,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "writing"),
				fmt.Sprintf("requests information field %s.", "writing"),
				[]string{"podname", "httpport"}, nil,
			),
			prometheus.GaugeValue, float64(writing), containername, httpport,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "nginx_status", "waiting"),
				fmt.Sprintf("requests information field %s.", "waiting"),
				[]string{"podname", "httpport"}, nil,
			),
			prometheus.GaugeValue, float64(waiting), containername, httpport,
		)

	}
	return nil

}
