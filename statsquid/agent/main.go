package agent

import (
	"fmt"
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
	verbose        bool
	needsSync      bool
	agentStats     *AgentStats
}

type AgentStats struct {
	counter    int64
	lastReport time.Time
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

	return &Agent{
		dockerClient:   dockerClient,
		mantleClient:   mantleClient,
		statQ:          newStatQ(),
		nodeInfo:       models.NewNode(info),
		nodeContainers: make(models.ContainerMap),
		collectors:     make(map[string]*Collector),
		verbose:        opts.Verbose,
		needsSync:      true,
		agentStats:     &AgentStats{0, time.Now()},
	}
}

func (agent *Agent) Run() {
	go agent.EventWatcher()
	go agent.SyncMantle()
	go agent.FlushStats()
	select {}
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

//Flush queued stats to mantle
func (agent *Agent) FlushStats() {
	var err error
	var reply int
	if !agent.statQ.isEmpty() {
		err = agent.mantleClient.Call("GiantAxon.FlushToMantle", agent.statQ.flush(), &reply)
		util.FailOnError(err)
	} else {
		time.Sleep(1 * time.Second)
	}
	agent.FlushStats()
}

func (agent *Agent) SyncMantle() {
	if agent.needsSync {
		agent.ReportContainers()
	}
	agent.SyncCollectors()
	if agent.verbose && time.Since(agent.agentStats.lastReport).Seconds() > 60 {
		agent.report()
	}
	time.Sleep(1 * time.Second)
	agent.SyncMantle()
}

func (agent *Agent) report() {
	diff := strconv.FormatFloat(time.Since(agent.agentStats.lastReport).Seconds(), 'f', 3, 64)
	util.Output(fmt.Sprintf("%d active collectors", len(agent.collectors)))
	util.Output(fmt.Sprintf("%d stats collected in last %s seconds", agent.agentStats.counter, diff))
	agent.agentStats.counter = 0
	agent.agentStats.lastReport = time.Now()
}

func (agent *Agent) newCollector(containerID string) *Collector {
	if _, ok := agent.nodeContainers[containerID]; ok == false {
		util.Output(fmt.Sprintf("collector start failed. unknown container id: %s", containerID))
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
	util.Output(fmt.Sprintf("starting collector for container: %s", opts.ID))
	defer delete(agent.collectors, opts.ID)
	agent.dockerClient.Stats(opts)
	util.Output(fmt.Sprintf("stopping collector for container: %s", opts.ID))
}

//wrap stat with container metadata
func (agent *Agent) streamHandler(container *models.Container, stats chan *docker.Stats) {
	for stat := range stats {
		agent.statQ.add(&models.StatSquidStat{container, stat, &models.ProcessedStats{}})
		if agent.verbose {
			agent.agentStats.counter++
		}
	}
}
