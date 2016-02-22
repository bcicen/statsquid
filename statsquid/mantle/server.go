package mantle

import (
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
)

//Agent communication object
type GiantAxon struct {
	nerveMap    *NerveMap
	statStream  chan *models.StatSquidStat
	verbose     bool
	lastFlush   time.Time
	statCounter int
}

func (a *GiantAxon) FlushToMantle(data []byte, reply *int) error {
	stats := models.UnpackStats(data)
	a.statCounter += len(stats)
	for _, stat := range stats {
		a.statStream <- stat
	}
	if time.Since(a.lastFlush).Seconds() > 10 {
		diff := strconv.FormatFloat(time.Since(a.lastFlush).Seconds(), 'f', 3, 64)
		util.Output("%v", a.statCounter, "stats collected in last", diff, "seconds")
		a.statCounter = 0
		a.lastFlush = time.Now()
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
	statStream := make(chan *models.StatSquidStat)
	axon := &GiantAxon{
		nerveMap:   newNerveMap(opts.Verbose),
		statStream: statStream,
		verbose:    opts.Verbose,
		lastFlush:  time.Now(),
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
