package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

//Agent communication object
type GiantAxon struct {
	statStream chan string
}

func (t *GiantAxon) AppendStat(stat string, reply *int) error {
	t.statStream <- stat
	*reply = 1
	return nil
}

func (t *GiantAxon) readOut() {
	var stat StatSquidStat
	for s := range t.statStream {
		stat.Unpack(s)
		j, err := json.Marshal(stat)
		failOnError(err)
		fmt.Println(string(j))
	}
}

func mantleServer(listenPort int) {
	axon := &GiantAxon{
		statStream: make(chan string),
	}
	rpc.Register(axon)
	rpc.HandleHTTP()
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(listenPort))
	failOnError(err)

	go http.Serve(listen, nil)
	axon.readOut()
}
