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
	var msg string
	var lastmsg string

	defer ws.Close()
	w.clients = append(w.clients, ws)
	if w.verbose {
		util.Output(fmt.Sprintf("client connected to websocket: %s", ws.Request().RemoteAddr))
	}

	for {
		websocket.Message.Receive(ws, &msg)
		if msg == lastmsg {
			//TODO: remove websocket connection object from WebSocketServer.clients
			if w.verbose {
				util.Output(fmt.Sprintf("websocket client disconnected: %s", ws.Request().RemoteAddr))
			}
			return
		}
		lastmsg = msg
		time.Sleep(5 * time.Second)
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
			util.Output(fmt.Sprintf("%d stats collected in last %s seconds", a.statCounter, diff))
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

	util.Output(fmt.Sprintf("mantle server listening on :%d", opts.ListenPort))
	http.ListenAndServe("0.0.0.0:"+strconv.Itoa(opts.ListenPort), nil)
}
