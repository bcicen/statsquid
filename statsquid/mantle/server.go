package mantle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"github.com/vektorlab/statsquid/models"
	"github.com/vektorlab/statsquid/util"
	"golang.org/x/net/websocket"
)

type WebSocketServer struct {
	clients []*websocket.Conn
	stream  chan []byte
	verbose bool
}

func newWebSocketServer(verbose bool) *WebSocketServer {
	w := &WebSocketServer{
		clients: make([]*websocket.Conn, 0),
		stream:  make(chan []byte),
		verbose: verbose,
	}
	go w.clientStream()
	return w
}

func (w *WebSocketServer) handler(ws *websocket.Conn) {
	if w.verbose {
		fmt.Println("client connected to websocket: %s", ws.Request().RemoteAddr)
	}
	var msg string
	w.clients = append(w.clients, ws)
	for {
		websocket.Message.Receive(ws, &msg)
		fmt.Println(msg)
	}
}

func (w *WebSocketServer) clientStream() {
	for s := range w.stream {
		for _, c := range w.clients {
			websocket.Message.Send(c, s)
		}
	}
}

//Agent communication object
type GiantAxon struct {
	nerveMap    *NerveMap
	wsServer    *WebSocketServer
	verbose     bool
	lastFlush   time.Time
	statCounter int
}

func (a *GiantAxon) FlushToMantle(data []byte, reply *int) error {
	stats := models.UnpackStats(data)
	for _, stat := range stats {
		a.nerveMap.updateStat(stat)
		j, err := json.Marshal(stat)
		util.FailOnError(err)
		a.wsServer.stream <- j
	}
	if a.verbose {
		a.statCounter += len(stats)
		if time.Since(a.lastFlush).Seconds() > 10 {
			diff := strconv.FormatFloat(time.Since(a.lastFlush).Seconds(), 'f', 3, 64)
			util.Output("%v", a.statCounter, "stats collected in last", diff, "seconds")
			a.statCounter = 0
			a.lastFlush = time.Now()
		}
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
	ListenPort int
	Verbose    bool
}

func MantleServer(opts *MantleServerOpts) {
	axon := &GiantAxon{
		nerveMap:  newNerveMap(opts.Verbose),
		wsServer:  newWebSocketServer(opts.Verbose),
		verbose:   opts.Verbose,
		lastFlush: time.Now(),
	}
	//init RPC server
	rpc.Register(axon)
	rpc.HandleHTTP()
	http.Handle("/ws", websocket.Handler(axon.wsServer.handler))

	util.Output("mantle server listening on :%d", opts.ListenPort)
	http.ListenAndServe("0.0.0.0:"+strconv.Itoa(opts.ListenPort), nil)
}
