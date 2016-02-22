package models

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/vektorlab/statsquid/util"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type StatSquidStat struct {
	*Container
	*docker.Stats
}

func PackStats(s []*StatSquidStat) []byte {
	packed, err := msgpack.Marshal(s)
	if err != nil {
		util.Output("stat marshal failed: %s", err)
		return nil
	}
	return packed
}

func UnpackStats(b []byte) []*StatSquidStat {
	var stats []*StatSquidStat
	err := msgpack.Unmarshal(b, &stats)
	if err != nil {
		util.Output("stat unmarshal failed: %s", err)
		return nil
	}
	return stats
}
