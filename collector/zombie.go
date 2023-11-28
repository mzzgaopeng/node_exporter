package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/procfs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	zombieProcessesInfo = "zombie_processes_info"
)

type newZombieInfoCollector struct{}

func init() {
	registerCollector("storagemapper-k8s", true, NewZombieInfoCollector)
}

// NewContainerdCollector returns a new Collector.
func NewZombieInfoCollector() (Collector, error) {
	return &newZombieInfoCollector{}, nil
}

// 处理docker返回的数据 并筛选出有僵尸进程的
func updateZombieInfo(mapperList []string, mapper *[]string) {
	zombiePidAndContainerPid, err := GetZombiePidAndContainerPid()
	if err != nil {
		log.Error("GetZombiePidAndContainerPid： %q", err)
	}
	for _, mapperItem := range mapperList {
		if !strings.Contains(mapperItem, "no value") && mapperItem != "" {
			mapperKV := strings.Split(mapperItem, "__")
			name := mapperKV[0]
			containerPid := mapperKV[1]
			//筛选出僵尸进程
			_, ok := zombiePidAndContainerPid[containerPid]
			if ok {
				*mapper = append(*mapper, name)
			}
		}
	}
}

func (c *newZombieInfoCollector) Update(ch chan<- prometheus.Metric) error {

	var mapper []string

	// 执行docker命令 获取pid和容器名称的对应关系
	pidName, err := execCommand("docker inspect --format='{{.Name}}__{{.State.Pid}}' $(docker ps -aq)")
	if err != nil {
		log.Debugf("Get 容器Pid fail: %q", err)
	}

	updateZombieInfo(strings.Split(string(pidName), "\n"), &mapper)

	for _, containerName := range mapper {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, zombieProcessesInfo, "PidMapper"),
				fmt.Sprintf("zombie containerName information"),
				[]string{"Name", "Pid"}, nil,
			),
			prometheus.GaugeValue, float64(1), containerName,
		)
	}

	return nil
}

// 获取所有僵尸进程和容器启动进程的map
func GetZombiePidAndContainerPid() (map[string]int, error) {
	fs, err := procfs.NewFS(*procPath)
	if err != nil {
		return nil, err
	}
	p, err := fs.AllProcs()
	if err != nil {
		return nil, err
	}
	//获取进程和父进程的对应关系
	processParents := make(map[int]int)
	for _, pid := range p {
		stat, err := pid.NewStat()
		// PIDs can vanish between getting the list and getting stats.
		if os.IsNotExist(err) {
			log.Debugf("file not found when retrieving stats: %q", err)
			continue
		}
		if err != nil {
			return nil, err
		}
		processParents[stat.PID] = stat.PPID
	}
	//获取所有僵尸进程的容器启动进程切片
	var zombiePid int
	containerPids := make(map[string]int)
	for _, pid := range p {
		stat, err := pid.NewStat()
		// PIDs can vanish between getting the list and getting stats.
		if os.IsNotExist(err) {
			log.Debugf("file not found when retrieving stats: %q", err)
			continue
		}
		if err != nil {
			return nil, err
		}

		if stat.State == "Z" {
			zombiePid = stat.PID
			for i := 0; i < 10; i++ {
				_, sign, err := GetCmdlineByPid(zombiePid)
				if err != nil {
					log.Error("GetCmdlineByPid： %q", err)
					continue
				}
				if sign {
					containerPids[strconv.Itoa(zombiePid)] = stat.PID
					break
				} else {
					zombiePid = processParents[zombiePid]
				}
			}
		}
	}
	return containerPids, nil
}

// 根据进程号获取cmdline 并判断cmdline中是否包含容器启动命令 包含就返回true
func GetCmdlineByPid(pid int) (string, bool, error) {
	cmdlinePath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "cmdline")

	cmdlineBytes, err := os.ReadFile(cmdlinePath)
	if err != nil {
		fmt.Printf("无法读取进程 %d 的cmdline信息：%v\n", pid, err)
		return "", false, err
	}
	cmdline := string(cmdlineBytes)
	//判断cmdline是否包含containerd-shim 包含就返回true
	if strings.Contains(cmdline, "containerd-shim") {
		fmt.Printf("进程 %d 的cmdline信息：%s\n", pid, cmdline)
		return cmdline, true, nil
	}
	fmt.Printf("进程 %d 的cmdline信息：%s\n", pid, cmdline)
	return cmdline, false, nil
}
