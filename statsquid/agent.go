package main

import (
	"time"

	"github.com/fsouza/go-dockerclient"
	"gopkg.in/redis.v3"
)

type Agent struct {
	docker     *docker.Client
	redis      *redis.Client
	allStats   chan string
	collectors map[string]*Collector
	hostInfo   map[string]string
}

type Collector struct {
	stats chan *docker.Stats
	done  chan bool
	opts  docker.StatsOptions
}

func newAgent(dockerHost string) *Agent {
	api, err := docker.NewClient(dockerHost)
	failOnError(err)
	info, err := api.Info()
	failOnError(err)

	redis := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	_, err = redis.Ping().Result()
	failOnError(err)

	agent := &Agent{
		docker:     api,
		redis:      redis,
		allStats:   make(chan string),
		collectors: make(map[string]*Collector),
		hostInfo:   info.Map(),
	}
	return agent
}

func (agent *Agent) watchContainers(verbose bool) {
	containers, err := agent.docker.ListContainers(docker.ListContainersOptions{})
	failOnError(err)
	for _, c := range containers {
		if _, ok := agent.collectors[c.ID]; ok == false {
			container := agent.newContainer(c)
			agent.collectors[c.ID] = agent.newCollector(container)
		}
	}
	if verbose {
		agent.report()
	}
	time.Sleep(1 * time.Second)
	agent.watchContainers(verbose)
}

func (agent *Agent) report() {
	output("%d active collectors", len(agent.collectors))
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
		agent.allStats <- packedStat
	}
}

func (agent *Agent) streamOut() {
	for s := range agent.allStats {
		err := agent.redis.Publish("statsquid", s).Err()
		failOnError(err)
	}
}
