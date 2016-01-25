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
				agent := newAgent(c.String("docker-host"))
				go agent.watchContainers(c.GlobalBool("verbose"))
				go agent.streamOut()
				select {}
			},
		},
		{
			Name:  "listener",
			Usage: "statsquid listener",
			Action: func(c *cli.Context) {
				readIn()
			},
		},
	}

	app.Run(os.Args)
}
