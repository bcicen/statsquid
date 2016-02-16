package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/vektorlab/statsquid/mantle"
)

var version = "dev-build"

func main() {
	app := cli.NewApp()
	app.Name = "statsquid"
	app.Usage = "Container stats aggregator"
	app.Version = version

	//global flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Verbose output",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "agent",
			Usage: "statsquid agent",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "docker-host, d",
					Value:  "unix:///var/run/docker.sock",
					Usage:  "Docker host",
					EnvVar: "DOCKER_HOST",
				},
				cli.StringFlag{
					Name:   "mantle-host, m",
					Value:  "127.0.0.1:1234",
					Usage:  "Mantle host",
					EnvVar: "MANTLE_HOST",
				},
			},
			Action: func(c *cli.Context) {
				agent := newAgent(&AgentOpts{
					mantleHost: c.String("mantle-host"),
					dockerHost: c.String("docker-host"),
					verbose:    c.GlobalBool("verbose"),
				})
				go agent.syncMantle()
				go agent.streamOut()
				select {}
			},
		},
		{
			Name:  "mantle",
			Usage: "statsquid mantle",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:   "listen, l",
					Value:  1234,
					Usage:  "Port to listen on",
					EnvVar: "MANTLE_PORT",
				},
				cli.StringFlag{
					Name:   "elastic-host",
					Value:  "127.0.0.1",
					Usage:  "Elasticsearch host",
					EnvVar: "MANTLE_ES_HOST",
				},
			},
			Action: func(c *cli.Context) {
				opts := &mantle.MantleServerOpts{
					ListenPort:  c.Int("listen"),
					ElasticHost: c.String("elastic-host"),
					ElasticPort: 9300,
					Verbose:     c.GlobalBool("verbose"),
				}
				mantle.MantleServer(opts)
			},
		},
	}

	app.Run(os.Args)
}
