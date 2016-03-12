package models

import (
	"strconv"

	"github.com/fsouza/go-dockerclient"
)

type ReportContainersMsg struct {
	Node       *Node
	Containers ContainerMap
}

type Node struct {
	NodeID       string
	NodeName     string
	NodeNcpu     float64
	NodeMemTotal float64
}

func NewNode(env *docker.Env) *Node {
	envMap := env.Map()
	ncpu, err := strconv.ParseFloat(envMap["NCPU"], 64)
	if err != nil {
		ncpu = float64(0)
	}
	nmem, err := strconv.ParseFloat(envMap["MemTotal"], 64)
	if err != nil {
		nmem = float64(0)
	}
	return &Node{
		NodeID:       envMap["ID"],
		NodeName:     envMap["Name"],
		NodeNcpu:     ncpu,
		NodeMemTotal: nmem,
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
