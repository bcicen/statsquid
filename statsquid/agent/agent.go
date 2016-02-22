package agent

import (
	"net/rpc"
	"strconv"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
)

type AgentOpts struct {
	MantleHost string
	DockerHost string
	Verbose    bool
}

type Agent struct {
	dockerClient   *docker.Client
	mantleClient   *rpc.Client
	statQ          *StatQ
	nodeInfo       *models.Node
	nodeContainers models.ContainerMap
	collectors     map[string]*Collector
	counter        int64
	verbose        bool
	needsSync      bool
	lastReport     time.Time
}

type Collector struct {
	stats chan *docker.Stats
	done  chan bool
	opts  docker.StatsOptions
}

func NewAgent(opts *AgentOpts) *Agent {
	dockerClient, err := docker.NewClient(opts.DockerHost)
	util.FailOnError(err)

	mantleClient, err := rpc.DialHTTP("tcp", opts.MantleHost)
	util.FailOnError(err)

	info, err := dockerClient.Info()
	util.FailOnError(err)

	agent := &Agent{
		dockerClient:   dockerClient,
		mantleClient:   mantleClient,
		statQ:          newStatQ(),
		nodeInfo:       models.NewNode(info),
		nodeContainers: make(models.ContainerMap),
		collectors:     make(map[string]*Collector),
		verbose:        opts.Verbose,
		needsSync:      true,
		lastReport:     time.Now(),
	}
	go agent.EventWatcher()
	return agent
}

//Trigger ReportContainers on specified Docker events
func (agent *Agent) EventWatcher() {
	events := make(chan *docker.APIEvents)
	agent.dockerClient.AddEventListener(events)
	for e := range events {
		if e.Status == "start" || e.Status == "die" {
			agent.needsSync = true
		}
	}
}

//Report running containers to Mantle
func (agent *Agent) ReportContainers() {
	var reply *int

	report := &models.ReportContainersMsg{
		Node:       agent.nodeInfo,
		Containers: make(models.ContainerMap),
	}
	containers, err := agent.dockerClient.ListContainers(docker.ListContainersOptions{})
	util.FailOnError(err)
	for _, c := range containers {
		report.Containers[c.ID] = models.NewContainer(agent.nodeInfo, c)
	}

	err = agent.mantleClient.Call("GiantAxon.ReportContainers", report, &reply)
	util.FailOnError(err)

	agent.needsSync = false
	agent.nodeContainers = report.Containers
	if agent.verbose {
		util.Output("synced running containers to mantle")
	}
}

func (agent *Agent) SyncCollectors() {
	var collectors map[string]bool

	err := agent.mantleClient.Call("GiantAxon.GetCollectors", agent.nodeInfo.NodeID, &collectors)
	util.FailOnError(err)

	for id, active := range collectors {
		//start collectors marked active
		if active {
			if _, ok := agent.collectors[id]; ok == false {
				collector := agent.newCollector(id)
				if collector != nil {
					agent.collectors[id] = collector
				}
			}
		} else {
			//stop collectors not marked active
			if _, ok := agent.collectors[id]; ok == true {
				agent.collectors[id].done <- true
			}
		}
	}
}

func (agent *Agent) SyncMantle() {
	if agent.needsSync {
		agent.ReportContainers()
	}
	agent.SyncCollectors()
	if agent.verbose && time.Since(agent.lastReport).Seconds() > 10 {
		agent.report()
	}
	time.Sleep(1 * time.Second)
	agent.SyncMantle()
}

func (agent *Agent) report() {
	diff := strconv.FormatFloat(time.Since(agent.lastReport).Seconds(), 'f', 3, 64)
	util.Output("%d active collectors", len(agent.collectors))
	util.Output("%v", agent.counter, "stats collected in last", diff, "seconds")
	agent.counter = 0
	agent.lastReport = time.Now()
}

func (agent *Agent) newCollector(containerID string) *Collector {
	if _, ok := agent.nodeContainers[containerID]; ok == false {
		util.Output("collector start failed. unknown container id: %s", containerID)
		return nil
	}

	exitChannel := make(chan bool)
	statsChannel := make(chan *docker.Stats)
	container := agent.nodeContainers[containerID]

	collector := &Collector{
		stats: statsChannel,
		done:  exitChannel,
		opts: docker.StatsOptions{
			ID:     container.ID,
			Stats:  statsChannel,
			Stream: true,
			Done:   exitChannel,
		},
	}

	go agent.collect(collector.opts)
	go agent.streamHandler(container, statsChannel)
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
func (agent *Agent) streamHandler(container *models.Container, stats chan *docker.Stats) {
	for stat := range stats {
		agent.statQ.add(&models.StatSquidStat{container, stat})
		if agent.verbose {
			agent.counter++
		}
	}
}

func (agent *Agent) StreamOut() {
	var err error
	var reply int
	for {
		if !agent.statQ.isEmpty() {
			err = agent.mantleClient.Call("GiantAxon.FlushToMantle", agent.statQ.flush(), &reply)
			util.FailOnError(err)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
