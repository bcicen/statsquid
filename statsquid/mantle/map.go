package mantle

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
)

//NerveMap holds the global state for all reporting agents
type NerveMap struct {
	nodeMap      map[string]models.ContainerMap   //mapping of node id to list of containers
	collectorMap map[string]bool                  //mapping of active collectors
	statMap      map[string]*models.StatSquidStat //mapping of container id to most recent stat
	lock         sync.RWMutex
	verbose      bool
}

func newNerveMap(verbose bool) *NerveMap {
	n := &NerveMap{
		nodeMap:      make(map[string]models.ContainerMap),
		collectorMap: make(map[string]bool),
		statMap:      make(map[string]*models.StatSquidStat),
		lock:         sync.RWMutex{},
		verbose:      verbose,
	}
	go n.cleanupCollectors()
	return n
}

func (m *NerveMap) statMapToJSON() []byte {
	j, err := json.Marshal(m.statMap)
	util.FailOnError(err)
	return j
}

func (m *NerveMap) updateStat(stat *models.StatSquidStat) {
	cid := stat.ID
	if _, ok := m.statMap[cid]; ok == true {
		models.CalculateCPU(m.statMap[cid], stat)
		models.CalculateNet(stat)
		models.CalculateBlkIO(stat)
	}
	m.statMap[cid] = stat
}

//regulary remove IDs not matching any registered containers
func (m *NerveMap) cleanupCollectors() {
	for {
		m.lock.Lock()
		for id := range m.collectorMap {
			if !m.containerExists(id) {
				delete(m.collectorMap, id)
				delete(m.statMap, id)
				util.Output(fmt.Sprintf("removed stale collector toggle: %s", id))
			}
		}
		m.lock.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func (m *NerveMap) containerExists(id string) bool {
	for _, cmap := range m.nodeMap {
		if _, ok := cmap[id]; ok == true {
			return true
		}
	}
	return false
}

func (m *NerveMap) collectorExists(id string) bool {
	_, ok := m.collectorMap[id]
	return ok
}

func (m *NerveMap) updateNodeContainers(report *models.ReportContainersMsg) {
	m.lock.Lock()
	//create a collector toggle for new containers
	for id := range report.Containers {
		if !m.collectorExists(id) {
			m.collectorMap[id] = false
		}
	}
	//update our container map for given node
	nodeID := report.Node.NodeID
	if _, ok := m.nodeMap[nodeID]; ok == false {
		util.Output(fmt.Sprintf("New node registered: %s", nodeID))
	}
	m.nodeMap[nodeID] = report.Containers
	m.lock.Unlock()
}

func (m *NerveMap) getContainerByID(id string) *models.Container {
	for _, cmap := range m.nodeMap {
		if _, ok := cmap[id]; ok == true {
			return cmap[id]
		}
	}
	return nil
}

func (m *NerveMap) getContainersByNode(nodeID string) models.ContainerMap {
	return m.nodeMap[nodeID]
}

func (m *NerveMap) toggleAllCollectors(active bool) {
	for id := range m.collectorMap {
		if m.collectorMap[id] != active {
			m.collectorMap[id] = active
			if m.verbose {
				util.Output(fmt.Sprintf("collector toggled for %s", id))
			}
		}
	}
}

func (m *NerveMap) toggleCollector(containerID string) {
	container := m.getContainerByID(containerID)
	if container == nil {
		return
	}
	m.collectorMap[containerID] = (m.collectorMap[containerID] != true)
	if m.verbose {
		util.Output(fmt.Sprintf("collector toggled for %s on node %s", containerID, container.NodeName))
	}
}

func (m *NerveMap) getCollectorsByNode(nodeID string) map[string]bool {
	collectors := make(map[string]bool)
	for id := range m.getContainersByNode(nodeID) {
		collectors[id] = m.collectorMap[id]
	}
	return collectors
}
