package main

import (
	"os"

	"github.com/praveensastry/cm/terminal"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "cm"
	app.Usage = "Command Line Configuration Management System"
	app.Version = "0.0.1"
	app.Author = "Praveen Sastry"
	app.Email = "sastry.praveen@gmail.com"
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:        "list-hosts",
			ShortName:   "lh",
			Usage:       "cm list-hosts",
			Description: "List all hosts that are registered with cm",
			Action: func(c *cli.Context) error {

				return nil
			},
		},
		{
			Name:        "configure",
			ShortName:   "c",
			Usage:       "cm configure <spec>",
			Description: "Configure one or many remote servers with a given spec",
			Action: func(c *cli.Context) error {

				return nil
			},
		},
		{
			Name:        "add-host",
			ShortName:   "ah",
			Usage:       "cm add-host",
			Description: "Register a new host with cm",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:        "delete-host",
			ShortName:   "dh",
			Usage:       "cm delete-host",
			Description: "Deregister a host from cm",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:        "list-specs",
			ShortName:   "ls",
			Usage:       "cm list-specs",
			Description: "List all available specs",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:        "describe-spec",
			ShortName:   "ds",
			Usage:       "cm describe-spec",
			Description: "Show what a given spec will build",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}
	terminal.Information("buh")

	app.Run(os.Args)
}

func getConfig() {
	terminal.Information("buh")
}
