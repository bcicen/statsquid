package main

import (
	"net/rpc"
	"strconv"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
)

type AgentOpts struct {
	mantleHost string
	dockerHost string
	verbose    bool
}

type Agent struct {
	dockerClient *docker.Client
	mantleClient *rpc.Client
	nodeInfo     map[string]string
	allStats     chan []byte
	collectors   map[string]*Collector
	counter      int64
	verbose      bool
	lastReport   time.Time
}

type Collector struct {
	stats chan *docker.Stats
	done  chan bool
	opts  docker.StatsOptions
}

func newAgent(opts *AgentOpts) *Agent {
	dockerClient, err := docker.NewClient(opts.dockerHost)
	util.FailOnError(err)

	mantleClient, err := rpc.DialHTTP("tcp", opts.mantleHost)
	util.FailOnError(err)

	info, err := dockerClient.Info()
	util.FailOnError(err)

	agent := &Agent{
		dockerClient: dockerClient,
		mantleClient: mantleClient,
		nodeInfo:     info.Map(),
		allStats:     make(chan []byte),
		collectors:   make(map[string]*Collector),
		verbose:      opts.verbose,
		lastReport:   time.Now(),
	}

	return agent
}

func (agent *Agent) syncContainers() []*models.Container {
	var reply []*models.Container
	var nContainers []*models.Container

	containers, err := agent.dockerClient.ListContainers(docker.ListContainersOptions{})
	util.FailOnError(err)
	for _, c := range containers {
		nContainers = append(nContainers, models.NewContainer(agent.nodeInfo["Name"], agent.nodeInfo["NCpu"], c))
	}

	err = agent.mantleClient.Call("GiantAxon.SyncContainers", nContainers, &reply)
	util.FailOnError(err)

	return reply
}

func (agent *Agent) syncMantle() {
	util.Output("sync: %s", time.Now())
	for _, c := range agent.syncContainers() {
		if c.Watch {
			//start collectors for requested containers
			if _, ok := agent.collectors[c.ID]; ok == false {
				agent.collectors[c.ID] = agent.newCollector(c)
			}
		} else {
			//stop collectors for requested containers
			if _, ok := agent.collectors[c.ID]; ok == true {
				agent.collectors[c.ID].done <- true
			}
		}
	}

	if agent.verbose && time.Since(agent.lastReport).Seconds() > 60 {
		agent.report()
	}
	time.Sleep(3 * time.Second)
	agent.syncMantle()
}

func (agent *Agent) report() {
	diff := strconv.FormatFloat(time.Since(agent.lastReport).Seconds(), 'f', 3, 64)
	util.Output("%d active collectors", len(agent.collectors))
	util.Output("%v", agent.counter, "stats collected in last", diff, "seconds")
	agent.counter = 0
	agent.lastReport = time.Now()
}

func (agent *Agent) newCollector(c *models.Container) *Collector {
	exitChannel := make(chan bool)
	statsChannel := make(chan *docker.Stats)

	collector := &Collector{
		stats: statsChannel,
		done:  exitChannel,
		opts: docker.StatsOptions{
			ID:     c.ID,
			Stats:  statsChannel,
			Stream: true,
			Done:   exitChannel,
		},
	}

	go agent.collect(collector.opts)
	go agent.pack(c, statsChannel)
	return collector
}

//collect stats for given container
func (agent *Agent) collect(opts docker.StatsOptions) {
	util.Output("starting collector for container: %s", opts.ID)
	defer delete(agent.collectors, opts.ID)
	agent.dockerClient.Stats(opts)
	util.Output("stopping collector for container: %s", opts.ID)
}

//encode stats to aggregate channel
func (agent *Agent) pack(container *models.Container, stats chan *docker.Stats) {
	for stat := range stats {
		ss := &models.StatSquidStat{container, stat}
		agent.allStats <- ss.Pack()
		if agent.verbose {
			agent.counter++
		}
	}
}

func (agent *Agent) streamOut() {
	var err error
	var reply int
	for s := range agent.allStats {
		err = agent.mantleClient.Call("GiantAxon.SendStat", s, &reply)
		util.FailOnError(err)
	}
}
