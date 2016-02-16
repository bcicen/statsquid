package mantle

import (
	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
)

//docker node info
type NodeInfo map[string]string

//NerveMap holds the global state for all reporting agents
type NerveMap struct {
	cmap    map[string]*models.Container
	nmap    map[string]NodeInfo
	verbose bool
}

func newNerveMap(verbose bool) *NerveMap {
	return &NerveMap{
		cmap:    make(map[string]*models.Container),
		nmap:    make(map[string]NodeInfo),
		verbose: verbose,
	}
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
