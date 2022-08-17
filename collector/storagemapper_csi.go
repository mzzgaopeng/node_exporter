package collector

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"os/exec"
	"strings"
	"time"
)

const (
	storageCsiInfo = "storage_csi_status"
)

type newStorageCsiCollector struct{}

func init() {
	registerCollector("storagecsi-k8s", defaultDisabled, NewStorageCsiCollector)
}

func NewStorageCsiCollector() (Collector, error) {
	return &newStorageCsiCollector{}, nil
}

func updateStorageCsiMapper(mapperList []string, mapper *map[string]string) {
	for _, mapperItem := range mapperList {
		if strings.Contains(mapperItem, "kubernetes.io~csi") == true && mapperItem != "" {
			mapperKV := strings.Split(mapperItem, "__")
			key := mapperKV[0]
			value := mapperKV[1]
			(*mapper)[key] = value
		}
	}
}

func execCommandCsi(command string) ([]byte, error) {
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

	if err != nil {
		fmt.Println("Non-zero exit code:", err)
	}

	return out, err
}

func (c *newStorageCsiCollector) Update(ch chan<- prometheus.Metric) error {
	mapper := make(map[string]string)

	hostConfig, err := execCommand("docker inspect --format='{{.Name}}__{{.HostConfig.Binds}}' $(docker ps -aq)")
	if err != nil {
		log.Debug("Get HostConfig fail: %q", err)
	}

	updateStorageCsiMapper(append(strings.Split(string(hostConfig), "\n")), &mapper)

	for mapperPodInfo, storagePath := range mapper {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, storageCsiInfo, "CsiMapper"),
				fmt.Sprintf("csi mapper information"),
				[]string{"storagePath", "csiPodInfo"}, nil,
			),
			prometheus.GaugeValue, float64(1), storagePath, mapperPodInfo,
		)
	}

	return nil
}
