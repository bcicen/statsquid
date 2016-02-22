package models

import (
	"github.com/fsouza/go-dockerclient"
)

type ReportContainersMsg struct {
	Node       *Node
	Containers ContainerMap
}

type Node struct {
	NodeID       string
	NodeName     string
	NodeNcpu     string
	NodeMemTotal string
}

func NewNode(env *docker.Env) *Node {
	envMap := env.Map()
	return &Node{
		NodeID:       envMap["ID"],
		NodeName:     envMap["Name"],
		NodeNcpu:     envMap["NCPU"],
		NodeMemTotal: envMap["MemTotal"],
	}
}

type ContainerMap map[string]*Container

type ContainerBase struct {
	ID    string
	Image string
	Names []string
}

type Container struct {
	*Node
	*ContainerBase
}

func NewContainer(node *Node, c docker.APIContainers) *Container {
	base := &ContainerBase{
		ID:    c.ID,
		Image: c.Image,
		Names: c.Names,
	}
	return &Container{node, base}
}
