package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

var version = "dev-build"

func output(s string, a ...interface{}) {
	bold := color.New(color.Bold).SprintFunc()
	msg := fmt.Sprintf(s, a)
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
	app.Name = "statsquid-agent"
	app.Usage = "Container stats aggregator"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Verbose output",
		},
		cli.StringFlag{
			Name:   "docker-host, d",
			Value:  "unix:///var/run/docker.sock",
			Usage:  "Docker host",
			EnvVar: "DOCKER_HOST",
		},
	}
	app.Action = func(c *cli.Context) {
		agent := newAgent(c.String("docker-host"))
		go agent.watchContainers(c.Bool("verbose"))
		go agent.streamOut()
		select {}
	}

	app.Run(os.Args)
}
