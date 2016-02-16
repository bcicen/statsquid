package models

import (
	"github.com/fsouza/go-dockerclient"
)

type Container struct {
	NodeName string
	NodeNcpu string
	ID       string
	Image    string
	Names    []string
	Watch    bool
}

func NewContainer(node string, nodeNcpu string, c docker.APIContainers) *Container {
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
