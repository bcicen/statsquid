package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

var version = "dev-build"

func output(s string, a ...interface{}) {
	msg := fmt.Sprintf(s, a)
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", bold("statsquid"), msg)
}

func failOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

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
				go agent.watchContainers()
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
			},
			Action: func(c *cli.Context) {
				mantleServer(c.Int("listen"))
			},
		},
	}

	app.Run(os.Args)
}
