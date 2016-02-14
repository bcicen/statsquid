package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

//Agent communication object
type GiantAxon struct {
	containerMap *ContainerMap
	statStream   chan string
	verbose      bool
}

func (a *GiantAxon) SendStat(stat string, reply *int) error {
	a.statStream <- stat
	*reply = 1
	return nil
}

func (a *GiantAxon) ToggleCollector(id string, reply *int) error {
	c := a.containerMap.getContainerById(id)
	if c == nil {
		*reply = 1
		return nil
	}
	c.Watch = (c.Watch != true)
	if a.verbose {
		output("collector toggled for %s on node %s", c.ID, c.NodeName)
	}
	*reply = 0
	return nil
}

//report running containers to mantle, replying with a container map
//with global mantle directives applied, such as watch
func (a *GiantAxon) SyncContainers(containers []*Container, reply *[]*Container) error {
	//add containers we haven't seen before
	for _, c := range containers {
		if !a.containerMap.exists(c.ID) {
			a.containerMap.addContainer(c)
			output("saw new container on node ", c.NodeName, ": ", c.ID)
		}
	}
	*reply = a.containerMap.getContainersByNode(containers[0].NodeName)
	return nil
}

func (t *GiantAxon) readOut() {
	var stat StatSquidStat
	for s := range t.statStream {
		stat.Unpack(s)
		j, err := json.Marshal(stat)
		failOnError(err)
		fmt.Println(string(j))
	}
}

type mantleServerOpts struct {
	listenPort int
	verbose    bool
}

func mantleServer(opts *mantleServerOpts) {
	axon := &GiantAxon{
		containerMap: newContainerMap(),
		statStream:   make(chan string),
		verbose:      opts.verbose,
	}
	rpc.Register(axon)
	rpc.HandleHTTP()
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(opts.listenPort))
	failOnError(err)

	output("mantle server listening on :%d", opts.listenPort)
	go http.Serve(listen, nil)
	axon.readOut()
}
