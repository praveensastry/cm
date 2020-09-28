package main

import (
	"fmt"
	"os"

	"github.com/praveensastry/cm/internal/config"
	"github.com/praveensastry/cm/internal/parser"
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
				cfg := getConfig()
				terminal.Information(fmt.Sprintf("There are [%d] remote servers configured currently", len(cfg.Servers)))
				cfg.Servers.PrintAllServerInfo()
				return nil
			},
		},
		{
			Name:        "configure",
			ShortName:   "c",
			Usage:       "cm configure <spec>",
			Description: "Configure one or many remote servers with a given spec",
			Action: func(c *cli.Context) error {
				specList, err := parser.GetSpecs()
				if err != nil {
					terminal.ShowErrorMessage("Error Reading Spec Files!", err.Error())
					return err
				}

				cfg := getConfig()
				cfg.Servers.RemoteConfigure(c.Args().Get(0), specList)
				return nil
			},
		},
		{
			Name:        "add-host",
			ShortName:   "ah",
			Usage:       "cm add-host",
			Description: "Register a new host with cm",
			Action: func(c *cli.Context) error {
				cfg := getConfig()
				return cfg.AddServer()
			},
		},
		{
			Name:        "delete-host",
			ShortName:   "dh",
			Usage:       "cm delete-host",
			Description: "Deregister a host from cm",
			Action: func(c *cli.Context) error {
				cfg := getConfig()
				return cfg.DeleteServer()
			},
		},
		{
			Name:        "list-specs",
			ShortName:   "ls",
			Usage:       "cm list-specs",
			Description: "List all available specs",
			Action: func(c *cli.Context) error {
				specList, err := parser.GetSpecs()
				if err != nil {
					terminal.ShowErrorMessage("Error Reading Spec Files!", err.Error())
					return err
				}

				terminal.Information(fmt.Sprintf("There are [%d] specs available currently", len(specList.Specs)))
				specList.PrintSpecInformation()

				return nil
			},
		},
		{
			Name:        "describe-spec",
			ShortName:   "ds",
			Usage:       "cm describe-spec",
			Description: "Show what a given spec will build",
			Action: func(c *cli.Context) error {
				specList, err := parser.GetSpecs()
				if err != nil {
					terminal.ShowErrorMessage("Error Reading Spec Files!", err.Error())
				}

				specName := c.Args().Get(0)
				terminal.Information(fmt.Sprintf("Showing spec plan for spec: [%s]", specName))
				if !specList.SpecExists(specName) {
					terminal.ShowErrorMessage("Unable to find Spec!", fmt.Sprintf("I was unable to find a spec named [%s].", specName))
					return nil
				}

				specList.ShowSpecBuild(specName)
				return nil
			},
		},
	}
	app.Run(os.Args)
}

func getConfig() *config.CMConfig {
	// Check Config
	cfg, err := config.ReadConfig()
	if err != nil || len(cfg.Servers) == 0 {
		// No Config Found, ask if we want to create one
		create := terminal.BoxPromptBool("configuration file not found or empty!", "Do you want to add some servers now?")
		if !create {
			terminal.Information("Alright then, maybe next time.. ")
			os.Exit(0)
			return nil
		}
		// Add Some Servers to our config
		cfg.AddServer()
		os.Exit(0)
		return nil
	}

	return cfg
}
