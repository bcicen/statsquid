package models

import (
	"fmt"

	"github.com/fsouza/go-dockerclient"
	"github.com/vektorlab/statsquid/util"
	"gopkg.in/vmihailenco/msgpack.v2"
)

const (
	cpuTick = 100
)

type StatSquidStat struct {
	*Container
	*docker.Stats
	*ProcessedStats
}

type ProcessedStats struct {
	CPUPercentage float64
}

func NewStatSquidStat(container *Container, stat *docker.Stats) {
}

//Calculate time-base CPU utilization from two stats
func CalculateCPU(prevStat, curStat *StatSquidStat) {
	timeDelta := curStat.Read.Sub(prevStat.Read).Seconds()
	systemDelta := float64(curStat.CPUStats.SystemCPUUsage -
		prevStat.CPUStats.SystemCPUUsage)
	containerDelta := float64(curStat.CPUStats.CPUUsage.TotalUsage -
		prevStat.CPUStats.CPUUsage.TotalUsage)
	// average usage deltas per second
	if timeDelta > 1 {
		systemDelta = systemDelta / timeDelta
		containerDelta = containerDelta / timeDelta
	}
	//for odd cases where system cpu is reportedly unchanged
	if systemDelta < 0 {
		curStat.CPUPercentage = 0
	}
	curStat.CPUPercentage = ((containerDelta / systemDelta) * cpuTick * curStat.NodeNcpu)
}

func PackStats(s []*StatSquidStat) []byte {
	packed, err := msgpack.Marshal(s)
	if err != nil {
		util.Output(fmt.Sprintf("stat marshal failed: %s", err))
		return nil
	}
	return packed
}

func UnpackStats(b []byte) []*StatSquidStat {
	var stats []*StatSquidStat
	err := msgpack.Unmarshal(b, &stats)
	if err != nil {
		util.Output(fmt.Sprintf("stat unmarshal failed: %s", err))
		return nil
	}
	return stats
}
