package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"os/exec"
	"strconv"
	"strings"
)

const (
	tunl0PortInfo = "tunl0_port"
)

type newTunl0PortCollector struct {}

func init() {
	registerCollector("tunl0port", true, NewTunl0PortCollector)
}

func NewTunl0PortCollector() (Collector, error) {
	return &newTunl0PortCollector{}, nil
}

/*func execShell(command string) ([]byte, error) {
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
}*/

func (c *newTunl0PortCollector) Update(ch chan<- prometheus.Metric) error {
	//ip, err := execShell("ifconfig tunl0 |grep inet|awk '{print $2}'")
	countPort := 1
	ip, err := exec.Command("/bin/bash", "-c", "ifconfig tunl0 |grep inet|awk '{print $2}'").Output()
	if err != nil {
		log.Debugf("Get tunl0 IP fail: %q", err)
	}
	if ip != nil {
		ipStr := strings.Replace(string(ip), "\n", "", -1)
		println(ipStr)
		shell := "netstat -anp |grep " + ipStr + " |grep ESTABLISHED|wc -l"
		count, err := exec.Command("/bin/bash", "-c", shell).Output()
		if err != nil {
			log.Debugf("Get tunl0 port count fail: %q", err)
		}
		countStr := strings.Replace(string(count), "\n", "", -1)
		countPort, _ = strconv.Atoi(string(countStr))
	} else {
		bondIP, err := exec.Command("/bin/bash", "-c", "ifconfig bond0 |grep inet|awk '{print $2}'").Output()
		bondIPStr := strings.Replace(string(bondIP), "\n", "", -1)
		shell := "netstat -anp |grep " + bondIPStr + " |grep ESTABLISHED|grep -E '172.23|172.24'|wc -l"
		count, err := exec.Command("/bin/bash", "-c", shell).Output()
		if err != nil {
			log.Debugf("Get tunl0 port count fail: %q", err)
		}
		countStr := strings.Replace(string(count), "\n", "", -1)
		countPort, _ = strconv.Atoi(string(countStr))
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, tunl0PortInfo, "count"),
			fmt.Sprintf("tunl0 port count"),
			nil, nil,
		),
		prometheus.CounterValue, float64(countPort),
	)
	return nil
}