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
			},
			Action: func(c *cli.Context) {
				transport, err := newTransport("127.0.0.1:6379")
				failOnError(err)
				agent := newAgent(&AgentOpts{
					dockerHost: c.String("docker-host"),
					verbose:    c.GlobalBool("verbose"),
				},
					transport)
				go agent.watchContainers()
				go agent.streamOut()
				select {}
			},
		},
		{
			Name:  "listener",
			Usage: "statsquid listener",
			Action: func(c *cli.Context) {
				transport, err := newTransport("127.0.0.1:6379")
				failOnError(err)
				readIn(transport)
			},
		},
	}

	app.Run(os.Args)
}
