package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	cmosEdgeCacheDirEnv = "EDGE_CACHE_DIR"

	cmosEdgeCollectorSubsystem  = "cmosedge"
	cmosEdgeProxyDefaultBaseDir = "/var/lib/kubeedge"
)

type edgeCollector struct {
	cacheList *prometheus.Desc
	cacheGet  *prometheus.Desc
}

func init() {
	registerCollector("cmosedge", defaultDisabled, NewEdgeCollector)
}

func NewEdgeCollector() (Collector, error) {
	return &edgeCollector{
		cacheList: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cmosEdgeCollectorSubsystem, "cacheList"),
			"边缘节点List缓存情况",
			[]string{"user_agent", "resource", "namespace"},
			nil,
		),
		cacheGet: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, cmosEdgeCollectorSubsystem, "cacheGet"),
			"边缘节点Get缓存情况",
			[]string{"user_agent", "resource", "namespace"},
			nil,
		),
	}, nil
}

func (c *edgeCollector) Update(ch chan<- prometheus.Metric) error {
	dir, _ := os.LookupEnv(cmosEdgeCacheDirEnv)
	if dir == "" {
		dir = cmosEdgeProxyDefaultBaseDir
	}

	for key, cnt := range c.cacheUpdate(filepath.Clean(dir + "/cache")) {
		ch <- prometheus.MustNewConstMetric(
			c.cacheList,
			prometheus.CounterValue,
			float64(cnt),
			strings.Split(key, "/")...,
		)
	}

	for key, cnt := range c.cacheUpdate(filepath.Clean(dir + "/cache-get")) {
		ch <- prometheus.MustNewConstMetric(
			c.cacheGet,
			prometheus.CounterValue,
			float64(cnt),
			strings.Split(key, "/")...,
		)
	}
	return nil
}

func (c *edgeCollector) cacheUpdate(dir string) map[string]int {
	var cacheCounterMap = map[string]int{}

	log.Debugf("start update cache list")
	fi1List, err := ioutil.ReadDir(dir)
	if err != nil {
		// TODO
		log.Errorf("read edge cache directory %s failed with error %+v", dir, err)
		return cacheCounterMap
	}

	// fi1 userAgent(dir)
	// fi2 namespace(dir) or resource name(file)
	// fi3 resource name(file)
	for _, fi1 := range fi1List {
		if !fi1.IsDir() {
			log.Warnf("edge cache directory %s not be valid cache", fi1.Name())
			continue
		}
		useragent := fi1.Name()
		fi2Path := filepath.Clean(dir + "/" + fi1.Name())
		fi2List, err := ioutil.ReadDir(fi2Path)
		if err != nil {
			log.Errorf("read edge cache directory %s failed with error %+v", fi2Path, err)
			continue
		}
		for _, fi2 := range fi2List {
			if !fi2.IsDir() {
				continue
			}
			gr := fi2.Name()

			fi3Path := filepath.Clean(fi2Path + "/" + fi2.Name())
			fi3List, err := ioutil.ReadDir(fi3Path)
			if err != nil {
				log.Errorf("read edge cache directory %s failed with error %+v", fi3Path, err)
				continue
			}

			for _, fi3 := range fi3List {
				if !fi3.IsDir() {
					cacheCounterMap[fmt.Sprintf("%s/%s/%s", useragent, gr, "")]++
					continue
				}
				ns := fi3.Name()
				fi4Path := filepath.Clean(fi3Path + "/" + fi3.Name())
				fi4List, err := ioutil.ReadDir(fi4Path)
				if err != nil {
					log.Errorf("read edge cache directory %s failed with error %+v", fi4Path, err)
					continue
				}
				for _, fi4 := range fi4List {
					if !fi4.IsDir() {
						cacheCounterMap[fmt.Sprintf("%s/%s/%s", useragent, gr, ns)]++
					}
				}
			}
		}
	}

	log.Debugf("dir: %s edgecache count %v", dir, cacheCounterMap)
	return cacheCounterMap
}
