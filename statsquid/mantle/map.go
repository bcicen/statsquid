package mantle

import (
	"time"

	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
)

//docker node info
type NodeInfo map[string]string

//NerveMap holds the global state for all reporting agents
type NerveMap struct {
	cmap    map[string]*models.Container
	nmap    map[string]NodeInfo
	ttlMap  map[string]time.Time
	verbose bool
}

func newNerveMap(verbose bool) *NerveMap {
	n := &NerveMap{
		cmap:    make(map[string]*models.Container),
		nmap:    make(map[string]NodeInfo),
		ttlMap:  make(map[string]time.Time),
		verbose: verbose,
	}
	go n.removeStaleContainers()
	return n
}

func (m *NerveMap) removeStaleContainers() {
	counter := 0
	for id, lastSeen := range m.ttlMap {
		if time.Since(lastSeen).Seconds() > 10 {
			m.delContainer(id)
			counter++
		}
	}
	util.Output("removed %d stale containers", counter)
	time.Sleep(5 * time.Second)
	m.removeStaleContainers()
}

func (m *NerveMap) containerExists(id string) bool {
	_, ok := m.cmap[id]
	return ok
}

func (m *NerveMap) nodeExists(id string) bool {
	_, ok := m.nmap[id]
	return ok
}

func (m *NerveMap) addContainer(c *models.Container) {
	m.cmap[c.ID] = c
	if m.verbose {
		util.Output("new container registered: %s %s", string(c.NodeName), string(c.ID))
	}
}

func (m *NerveMap) delContainer(id string) {
	delete(m.cmap, id)
	if m.verbose {
		util.Output("container de-registered: %s", id)
	}
}

func (m *NerveMap) bumpTTL(id string) {
	m.ttlMap[id] = time.Now()
}

func (m *NerveMap) addNode(n NodeInfo) {
	m.nmap[n["ID"]] = n
}

func (m *NerveMap) getContainerById(id string) *models.Container {
	if !m.containerExists(id) {
		return nil
	}
	return m.cmap[id]
}

func (m *NerveMap) getContainersByNode(node string) []*models.Container {
	var result []*models.Container
	for _, c := range m.cmap {
		if c.NodeName == node {
			result = append(result, c)
		}
	}
	return result
}
