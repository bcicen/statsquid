package mantle

import (
	"net"
	"net/http"
	"net/rpc"
	"strconv"

	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
)

//Agent communication object
type GiantAxon struct {
	nerveMap   *NerveMap
	statStream chan []byte
	verbose    bool
}

func (a *GiantAxon) SendStat(stat []byte, reply *int) error {
	a.statStream <- stat
	*reply = 1
	return nil
}

func (a *GiantAxon) ToggleAll(watch bool, reply *int) error {
	for _, c := range a.nerveMap.cmap {
		if c.Watch != watch {
			c.Watch = watch
			if a.verbose {
				util.Output("collector toggled for %s on node %s", c.ID, c.NodeName)
			}
		}
	}
	*reply = 0
	return nil
}

func (a *GiantAxon) ToggleCollector(id string, reply *int) error {
	c := a.nerveMap.getContainerById(id)
	if c == nil {
		*reply = 1
		return nil
	}
	c.Watch = (c.Watch != true)
	*reply = 0
	return nil
}

//report running containers to mantle, replying with a container map
//with global mantle directives applied, such as watch
func (a *GiantAxon) SyncContainers(containers []*models.Container, reply *[]*models.Container) error {
	reportingNode := containers[0].NodeName
	//add containers we haven't seen before
	for _, c := range containers {
		if !a.nerveMap.containerExists(c.ID) {
			a.nerveMap.addContainer(c)
		}
	}
	//remove stale containers
	var stillExists bool
	for _, c := range a.nerveMap.getContainersByNode(containers[0].NodeName) {
		for _, nc := range containers {
			if nc.ID == c.ID {
				stillExists = true
			}
		}
		if !stillExists {
			a.nerveMap.delContainer(c.ID)
		}
	}
	*reply = a.nerveMap.getContainersByNode(reportingNode)
	return nil
}

type MantleServerOpts struct {
	ListenPort  int
	ElasticHost string
	ElasticPort int
	Verbose     bool
}

func MantleServer(opts *MantleServerOpts) {
	statStream := make(chan []byte)
	axon := &GiantAxon{
		nerveMap:   newNerveMap(opts.Verbose),
		statStream: statStream,
		verbose:    opts.Verbose,
	}
	//init RPC server
	rpc.Register(axon)
	rpc.HandleHTTP()
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(opts.ListenPort))
	util.FailOnError(err)

	util.Output("mantle server listening on :%d", opts.ListenPort)
	go http.Serve(listen, nil)

	util.Output("starting indexer")
	siphon := newSiphon(opts.ElasticHost)
	siphon.worker(statStream)
}
