package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/procfs"
	"os"
	"path/filepath"
	"strings"
)

const (
	zombieProcessesInfo = "zombie_processes_info"
)

type newZombieInfoCollector struct{}

func init() {
	registerCollector("zombieProcessesInfo", true, NewZombieInfoCollector)
}

// NewZombieInfoCollector NewContainerdCollector returns a new Collector.
func NewZombieInfoCollector() (Collector, error) {
	return &newZombieInfoCollector{}, nil
}

// 根据容器id 查出容器名称
func updateZombieInfo(zombieContainerIds []string, mapper *map[string]string) {
	for _, zombieContainerId := range zombieContainerIds {
		// 执行docker命令入参是容器id 获取容器名称
		containerNameByte, err := execCommand("docker inspect --format='{{.Name}}' " + zombieContainerId)
		if err != nil {
			log.Debugf("Get 容器Pid fail: %q", err)
		}
		if string(containerNameByte) != "" {
			containerStr := strings.Split(string(containerNameByte), "_")
			(*mapper)[containerStr[2]] = containerStr[3]
		}
	}
}

func (c *newZombieInfoCollector) Update(ch chan<- prometheus.Metric) error {

	mapper := make(map[string]string)
	zombieContainerIds, err := GetZombieContainerIds()
	fmt.Printf("zombieContainerIds: %q\n", zombieContainerIds)
	if err != nil {
		log.Error("GetZombieContainerIds： %q", err)
	}

	updateZombieInfo(zombieContainerIds, &mapper)

	fmt.Println(mapper)

	for containerName, containerNamespace := range mapper {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, zombieProcessesInfo, "ContainerMapper"),
				fmt.Sprintf("zombie containerName information"),
				[]string{"containerName", "containerNamespace"}, nil,
			),
			prometheus.GaugeValue, float64(1), containerName, containerNamespace,
		)
	}

	return nil
}

// GetZombieContainerIds 获取所有僵尸进程的容器Id
func GetZombieContainerIds() ([]string, error) {
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
	var containerIds []string
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
				containerId, sign, err := GetContainerIdByPid(zombiePid)
				if err != nil {
					log.Error("GetContainerIdByPid： %q", err)
					continue
				}
				if sign {
					containerIds = append(containerIds, containerId)
					break
				} else {
					zombiePid = processParents[zombiePid]
				}
			}
		}
	}
	return containerIds, nil
}

// GetContainerIdByPid 根据进程号获取cmdline 并判断cmdline中是否包含容器启动命令 包含就返回容器ID和true
func GetContainerIdByPid(pid int) (string, bool, error) {
	cmdlinePath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "cmdline")

	cmdlineBytes, err := os.ReadFile(cmdlinePath)
	if err != nil {
		fmt.Printf("无法读取进程 %d 的cmdline信息：%v\n", pid, err)
		return "", false, err
	}
	cmdline := string(cmdlineBytes)
	//判断cmdline是否包含containerd-shim 包含就返回true
	var containerId string
	if strings.Contains(cmdline, "containerd-shim") {
		fmt.Printf("进程 %d 的cmdline信息：%s\n", pid, cmdline)
		//截取容器的ID 根据docker版本不同分两种情况
		if strings.Contains(cmdline, "docker-containerd-shim-namespacemoby-workdir") {
			containerId = cmdline[strings.LastIndex(cmdline, "moby/")+5 : strings.LastIndex(cmdline, "moby/")+17]
		} else if strings.Contains(cmdline, "containerd-shim-runc-v2-namespacemoby-id") {
			containerId = cmdline[strings.LastIndex(cmdline, "moby-id")+7 : strings.LastIndex(cmdline, "moby-id")+19]
		}
		return containerId, true, nil
	}
	fmt.Printf("进程 %d 的cmdline信息：%s\n", pid, cmdline)
	return containerId, false, nil
}
