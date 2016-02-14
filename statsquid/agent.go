package main

import (
	"net/rpc"
	"strconv"
	"time"

	"github.com/fsouza/go-dockerclient"
)

type AgentOpts struct {
	mantleHost string
	dockerHost string
	verbose    bool
}

type Agent struct {
	dockerClient *docker.Client
	mantleClient *rpc.Client
	allStats     chan string
	collectors   map[string]*Collector
	nodeInfo     map[string]string
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
	failOnError(err)

	mantleClient, err := rpc.DialHTTP("tcp", opts.mantleHost)
	failOnError(err)

	info, err := dockerClient.Info()
	failOnError(err)

	agent := &Agent{
		dockerClient: dockerClient,
		mantleClient: mantleClient,
		allStats:     make(chan string),
		collectors:   make(map[string]*Collector),
		nodeInfo:     info.Map(),
		verbose:      opts.verbose,
		lastReport:   time.Now(),
	}
	return agent
}

func (agent *Agent) syncContainers() []*Container {
	var reply []*Container
	var nContainers []*Container

	containers, err := agent.dockerClient.ListContainers(docker.ListContainersOptions{})
	failOnError(err)
	for _, c := range containers {
		nContainers = append(nContainers, newContainer(agent.nodeInfo["Name"], agent.nodeInfo["NCpu"], c))
	}

	err = agent.mantleClient.Call("GiantAxon.SyncContainers", nContainers, &reply)
	failOnError(err)

	return reply
}

func (agent *Agent) syncMantle() {
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

	if agent.verbose && time.Since(agent.lastReport).Seconds() > 10 {
		agent.report()
	}
	time.Sleep(1 * time.Second)
	agent.syncMantle()
}

func (agent *Agent) report() {
	diff := strconv.FormatFloat(time.Since(agent.lastReport).Seconds(), 'f', 3, 64)
	output("%d active collectors", len(agent.collectors))
	output("%v", agent.counter, "stats collected in last", diff, "seconds")
	agent.counter = 0
	agent.lastReport = time.Now()
}

func (agent *Agent) newCollector(c *Container) *Collector {
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
	output("starting collector for container: %s", opts.ID)
	agent.dockerClient.Stats(opts)
	output("stopping collector for container: %s", opts.ID)
	delete(agent.collectors, opts.ID)
}

//encode stats to aggregate channel
func (agent *Agent) pack(container *Container, stats chan *docker.Stats) {
	for stat := range stats {
		ss := &StatSquidStat{container, stat}
		packedStat, err := ss.Pack()
		failOnError(err)
		if agent.verbose {
			agent.counter++
		}
		agent.allStats <- packedStat
	}
}

func (agent *Agent) streamOut() {
	var err error
	var reply int
	for s := range agent.allStats {
		err = agent.mantleClient.Call("GiantAxon.SendStat", s, &reply)
		failOnError(err)
	}
}
