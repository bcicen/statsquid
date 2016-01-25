package main

import (
	"strconv"
	"time"

	"github.com/fsouza/go-dockerclient"
)

type AgentOpts struct {
	dockerHost string
	verbose    bool
}

type Agent struct {
	docker     *docker.Client
	transport  *Transport
	allStats   chan string
	collectors map[string]*Collector
	hostInfo   map[string]string
	counter    int64
	verbose    bool
	lastReport time.Time
}

type Collector struct {
	stats chan *docker.Stats
	done  chan bool
	opts  docker.StatsOptions
}

func newAgent(opts *AgentOpts, transport *Transport) *Agent {
	api, err := docker.NewClient(opts.dockerHost)
	failOnError(err)
	info, err := api.Info()
	failOnError(err)

	agent := &Agent{
		docker:     api,
		transport:  transport,
		allStats:   make(chan string),
		collectors: make(map[string]*Collector),
		hostInfo:   info.Map(),
		verbose:    opts.verbose,
		lastReport: time.Now(),
	}
	return agent
}

func (agent *Agent) watchContainers() {
	containers, err := agent.docker.ListContainers(docker.ListContainersOptions{})
	failOnError(err)
	for _, c := range containers {
		if _, ok := agent.collectors[c.ID]; ok == false {
			container := agent.newContainer(c)
			agent.collectors[c.ID] = agent.newCollector(container)
		}
	}
	if agent.verbose && time.Since(agent.lastReport).Seconds() > 10 {
		agent.report()
	}
	time.Sleep(1 * time.Second)
	agent.watchContainers()
}

func (agent *Agent) report() {
	diff := strconv.FormatFloat(time.Since(agent.lastReport).Seconds(), 'f', 3, 64)
	output("%d active collectors", len(agent.collectors))
	output("%v", agent.counter, "stats collected in last", diff, "seconds")
	agent.counter = 0
	agent.lastReport = time.Now()
}

func (agent *Agent) newContainer(c docker.APIContainers) *Container {
	container := &Container{
		Host:     agent.hostInfo["Name"],
		HostNcpu: agent.hostInfo["NCPU"],
		Id:       c.ID,
		Image:    c.Image,
		Names:    c.Names,
	}
	return container
}

func (agent *Agent) newCollector(c *Container) *Collector {
	exitChannel := make(chan bool)
	statsChannel := make(chan *docker.Stats)

	collector := &Collector{
		stats: statsChannel,
		done:  exitChannel,
		opts: docker.StatsOptions{
			ID:     c.Id,
			Stats:  statsChannel,
			Stream: true,
			Done:   make(chan bool),
		},
	}

	go agent.collect(collector.opts)
	go agent.pack(c, statsChannel)
	return collector
}

//collect stats for given container
func (agent *Agent) collect(opts docker.StatsOptions) {
	output("starting collector for container: %s", opts.ID)
	agent.docker.Stats(opts)
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
	for s := range agent.allStats {
		err := agent.transport.Publish(s)
		failOnError(err)
	}
}
