package main

import (
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type ContainerMap struct {
	cmap map[string]*Container
}

func newContainerMap() *ContainerMap {
	return &ContainerMap{
		cmap: make(map[string]*Container),
	}
}

func (m *ContainerMap) exists(id string) bool {
	_, ok := m.cmap[id]
	return ok
}

func (m *ContainerMap) addContainer(c *Container) {
	m.cmap[c.ID] = c
}

func (m *ContainerMap) getContainerById(id string) *Container {
	if !m.exists(id) {
		return nil
	}
	return m.cmap[id]
}

func (m *ContainerMap) getContainersByNode(node string) []*Container {
	var result []*Container
	for _, c := range m.cmap {
		if c.NodeName == node {
			result = append(result, c)
		}
	}
	return result
}

type Container struct {
	NodeName string
	NodeNcpu string
	ID       string
	Image    string
	Names    []string
	Watch    bool
}

func newContainer(node string, nodeNcpu string, c docker.APIContainers) *Container {
	container := &Container{
		NodeName: node,
		NodeNcpu: nodeNcpu,
		ID:       c.ID,
		Image:    c.Image,
		Names:    c.Names,
		Watch:    false,
	}
	return container
}

type StatSquidStat struct {
	*Container
	*docker.Stats
}

func (s *StatSquidStat) Pack() (string, error) {
	b, err := msgpack.Marshal(s)
	return string(b), err
}

func (s *StatSquidStat) Unpack(str string) {
	msgpack.Unmarshal([]byte(str), &s)
}
