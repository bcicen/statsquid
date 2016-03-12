package mantle

import (
	"time"

	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
)

//NerveMap holds the global state for all reporting agents
type NerveMap struct {
	nodeMap      map[string]models.ContainerMap   //mapping of node id to list of containers
	collectorMap map[string]bool                  //mapping of active collectors
	statMap      map[string]*models.StatSquidStat //mapping of container id to most recent stat
	verbose      bool
}

func newNerveMap(verbose bool) *NerveMap {
	n := &NerveMap{
		nodeMap:      make(map[string]models.ContainerMap),
		collectorMap: make(map[string]bool),
		statMap:      make(map[string]*models.StatSquidStat),
		verbose:      verbose,
	}
	go n.cleanupCollectors()
	return n
}

func (m *NerveMap) updateStat(stat *models.StatSquidStat) {
	cid := stat.ID
	if _, ok := m.statMap[cid]; ok == true {
		models.CalculateCPU(m.statMap[cid], stat)
	}
	m.statMap[cid] = stat
}

//regulary remove IDs in the collector map not matching any known containers
func (m *NerveMap) cleanupCollectors() {
	for id, _ := range m.collectorMap {
		if !m.containerExists(id) {
			delete(m.collectorMap, id)
			util.Output("removed stale collector toggle: %s", id)
		}
	}
	time.Sleep(30 * time.Second)
	m.cleanupCollectors()
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
	//create a collector toggle for new containers
	for id, _ := range report.Containers {
		if !m.collectorExists(id) {
			m.collectorMap[id] = false
		}
	}
	//update our container map for given node
	nodeID := report.Node.NodeID
	if _, ok := m.nodeMap[nodeID]; ok == false {
		util.Output("New node registered: %s", nodeID)
	}
	m.nodeMap[nodeID] = report.Containers
}

func (m *NerveMap) getContainerById(id string) *models.Container {
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
	for id, _ := range m.collectorMap {
		if m.collectorMap[id] != active {
			m.collectorMap[id] = active
			if m.verbose {
				util.Output("collector toggled for %s", id)
			}
		}
	}
}

func (m *NerveMap) toggleCollector(containerID string) {
	container := m.getContainerById(containerID)
	if container == nil {
		return
	}
	m.collectorMap[containerID] = (m.collectorMap[containerID] != true)
	if m.verbose {
		util.Output("collector toggled for %s on node %s", containerID, container.NodeName)
	}
}

func (m *NerveMap) getCollectorsByNode(nodeID string) map[string]bool {
	collectors := make(map[string]bool)
	for id, _ := range m.getContainersByNode(nodeID) {
		collectors[id] = m.collectorMap[id]
	}
	return collectors
}
