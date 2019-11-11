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
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	storageMapperInfo = "storage_mapper_status"
)

type newStorageMapperCollector struct{}

func init() {
	registerCollector("storagemapper-k8s", true, NewStorageMapperCollector)
}

// NewContainerdCollector returns a new Collector.
func NewStorageMapperCollector() (Collector, error) {
	return &newStorageMapperCollector{}, nil
}

func updateStorageMapper(mapperList []string, mapper *map[string]string) {
	for _, mapperItem := range mapperList {
		if !strings.Contains(mapperItem, "no value") && mapperItem != "" {
			mapperKV := strings.Split(mapperItem, "__")
			key := mapperKV[0]
			value := mapperKV[1]

			(*mapper)[key] = value
		}
	}
}

func execCommand(command string) ([]byte, error) {
	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // The cancel should be deferred so resources are cleaned up
	// Create the command with our context
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)

	// This time we can simply use Output() to get the result.
	out, err := cmd.Output()

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Command timed out")
		return nil, err
	}

	// If there's no context error, we know the command completed (or errored).
	fmt.Println("Output:", string(out))
	if err != nil {
		fmt.Println("Non-zero exit code:", err)
	}

	return out, err
}

// Update calls update containerd
func (c *newStorageMapperCollector) Update(ch chan<- prometheus.Metric) error {

	mapper := make(map[string]string)

	// DeviceMapper record
	devicemapper, err := execCommand("docker inspect --format='{{.Name}}__{{.GraphDriver.Data.DeviceName}}' $(docker ps -aq)")
	if err != nil {
		log.Debugf("Get devicemapper fail: %q", err)
	}
	// Overlay2 record
	overlay2, err := execCommand("docker inspect --format='{{.Name}}__{{.GraphDriver.Data.MergedDir}}' $(docker ps -aq)")
	if err != nil {
		log.Debugf("Get overlay2 fail: %q", err)
	}

	updateStorageMapper(append(strings.Split(string(devicemapper), "\n"),
		strings.Split(string(overlay2), "\n")...), &mapper)

	for mapperPodInfo, storagePath := range mapper {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, storageMapperInfo, "StorageMapper"),
				fmt.Sprintf("storage mapper information"),
				[]string{"storagePath", "mapperPodInfo"}, nil,
			),
			prometheus.GaugeValue, float64(1), storagePath, mapperPodInfo,
		)
	}

	return nil
}
