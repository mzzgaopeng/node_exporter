package collector

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func (c *rt6statCollector) getRt6stat() (map[string]float64, error) {
	r1, err := os.Open(procFilePath("net/rt6_stats"))
	if err != nil {
		return nil, err
	}
	defer r1.Close()

	r2, err := os.Open(procFilePath("sys/net/ipv6/route/max_size"))
	if err != nil {
		return nil, err
	}
	defer r2.Close()
	return parseRt6stat(r1, r2)
}

func parseRt6stat(r1, r2 io.Reader) (map[string]float64, error) {
	rt6Info := map[string]float64{}

	rt6statBytes, err := ioutil.ReadAll(r1)

	rt6useTmp := fmt.Sprintf("0x%s", strings.Fields(string(rt6statBytes))[5])
	rt6use, err := strconv.ParseInt(rt6useTmp, 0, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid value in meminfo: %s", err)
	}
	rt6Info["rt6_use"] = float64(rt6use)

	rt6totalBytes, err := ioutil.ReadAll(r2)
	rt6total := strings.Replace(string(rt6totalBytes), "\n", "", -1)
	rt6Info["rt6_total"], err = strconv.ParseFloat(rt6total, 64)

	return rt6Info, err
}
