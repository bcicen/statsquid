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

func (a *GiantAxon) SendStat(stats [][]byte, reply *int) error {
	util.Output("got %d stats", len(stats))
	for _, stat := range stats {
		a.statStream <- stat
	}
	*reply = 1
	return nil
}

func (a *GiantAxon) ToggleAll(active bool, reply *int) error {
	a.nerveMap.toggleAllCollectors(active)
	*reply = 0
	return nil
}

func (a *GiantAxon) ToggleCollector(id string, reply *int) error {
	a.nerveMap.toggleCollector(id)
	*reply = 0
	return nil
}

//report running containers to mantle
func (a *GiantAxon) ReportContainers(report *models.ReportContainersMsg, reply *int) error {
	a.nerveMap.updateNodeContainers(report)
	*reply = 1
	return nil
}

//return a map of containers to be watched for a given node
func (a *GiantAxon) GetCollectors(nodeID string, reply *map[string]bool) error {
	*reply = a.nerveMap.getCollectorsByNode(nodeID)
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
