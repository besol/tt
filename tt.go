package main

import (
	log "github.com/Sirupsen/logrus"
	log_hook "github.com/Sirupsen/logrus/hooks/sourcefile"
	"github.com/codegangsta/cli"
	"github.com/flexiant/tt/docker"
	"github.com/flexiant/tt/utils"
	"os"
	"path"
	"path/filepath"
)

const VERSION = "0.1.0"

func cmdNotFound(c *cli.Context, command string) {
	log.Fatalf(
		"%s: '%s' is not a %s command. See '%s --help'.",
		c.App.Name,
		command,
		c.App.Name,
		c.App.Name,
	)
}

func prepareFlags(c *cli.Context) error {
	if c.Bool("debug") {
		os.Setenv("DEBUG", "1")
		log.SetOutput(os.Stderr)
		log.SetLevel(log.DebugLevel)
		log.AddHook(&log_hook.SourceFileHook{LogLevel: log.InfoLevel})
	}
	os.Setenv("TT_ORIGIN", filepath.Clean(c.String("origin")))
	os.Setenv("TT_CONFIG", filepath.Clean(c.String("config")))
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Author = "Concerto Contributors"
	app.Email = "https://github.com/flexiant/tt"

	app.CommandNotFound = cmdNotFound
	app.Usage = "Wrapper to allow templating for Docker and Compose"
	app.Version = VERSION

	currentDir, err := os.Getwd()
	utils.CheckError(err)
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug mode",
		},
		cli.StringFlag{
			EnvVar: "TT_ORIGIN",
			Name:   "origin, o",
			Usage:  "Default Folder",
			Value:  currentDir,
		},
		cli.StringFlag{
			EnvVar: "TT_CONFIG",
			Name:   "config, c",
			Usage:  "Config Variables",
		},
	}

	app.Before = prepareFlags

	app.Commands = []cli.Command{
		{
			Name:   "docker",
			Usage:  "Manages docker with templating",
			Action: docker.Run,
		},
	}

	app.Run(os.Args)
}
