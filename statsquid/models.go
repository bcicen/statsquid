package main

import (
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Container struct {
	Host     string
	HostNcpu string
	Id       string
	Image    string
	Names    []string
}

type StatSquidStat struct {
	*Container
	*docker.Stats
}

func (s *StatSquidStat) Pack() (string, error) {
	b, err := msgpack.Marshal(s)
	return string(b), err
}

func (s *StatSquidStat) Unpack(str string) {
	msgpack.Unmarshal([]byte(str), &s)
}
