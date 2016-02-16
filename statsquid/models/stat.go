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

func (s *StatSquidStat) Pack() []byte {
	b, err := msgpack.Marshal(s)
	if err != nil {
		util.Output("stat marshal failed: %s", err)
		return nil
	}
	return b
}

func (s *StatSquidStat) Unpack(b []byte) {
	err := msgpack.Unmarshal(b, &s)
	if err != nil {
		util.Output("stat unmarshal failed: %s", err)
	}
}
