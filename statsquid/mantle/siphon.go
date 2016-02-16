package mantle

import (
	"encoding/json"
	"fmt"
	"time"

	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/vektorlab/statsquid/models"
)

//Elasticsearch client
type Siphon struct {
	indexer *elastigo.BulkIndexer
	done    chan bool
}

func newSiphon(elasticHost string) *Siphon {
	client := elastigo.NewConn()
	client.Domain = elasticHost
	siphon := &Siphon{
		indexer: client.NewBulkIndexerErrors(10, 60),
		done:    make(chan bool),
	}
	siphon.indexer.BufferDelayMax = (1 * time.Second)
	siphon.indexer.Start()
	go siphon.errHandler()
	return siphon
}

func (siphon *Siphon) worker(stream chan []byte) {
	var stat models.StatSquidStat
	for s := range stream {
		stat.Unpack(s)
		j, err := json.Marshal(stat)
		if err != nil {
			fmt.Println(err)
		} else {
			err = siphon.indexer.Index("stats", "stat", "", "", "", nil, j)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (siphon *Siphon) errHandler() {
	for errBuf := range siphon.indexer.ErrorChannel {
		fmt.Println(errBuf.Err)
	}
}
