package collector

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"os/exec"
	"strconv"
	"time"
)

const (
	tunl0PortInfo = "tunl0_port_count"
)

type newTunl0PortCollector struct {}

func init() {
	registerCollector("tunl0port", true, NewTunl0PortCollector)
}

func NewTunl0PortCollector() (Collector, error) {
	return &newTunl0PortCollector{}, nil
}

func execShell(command string) ([]byte, error) {
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

func (c *newTunl0PortCollector) Update(ch chan<- prometheus.Metric) error {
	ip, err := execShell("ifconfig tunl0 |grep inet|awk '{print $2}'")
	if err != nil {
		log.Debugf("Get tunl0 IP fail: %q", err)
	}

	count, err := execShell("netstat -anop |grep " + string(ip) + " |grep ESTABLISHED|wc -l")
	countPort, _ := strconv.Atoi(string(count))
	if err != nil {
		log.Debugf("Get tunl0 port count fail: %q", err)
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, tunl0PortInfo, "count"),
			fmt.Sprintf("tunl0 port count"),
			[]string{"portCount"}, nil,
		),
		prometheus.CounterValue, float64(countPort),
	)
	return nil
}