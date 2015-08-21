package docker

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/tt/containers"
	"github.com/flexiant/tt/utils"
	"io"
	"os"
	"os/exec"
	"strings"
)

func Run(c *cli.Context) {
	container := &containers.Container{}

	containerPwd, err := containers.PrepareContainer(container)
	utils.CheckError(err)

	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s; pwd; docker %s", containerPwd, strings.Join(c.Args(), " ")))

	stdout, err := cmd.StdoutPipe()
	utils.CheckError(err)

	stderr, err := cmd.StderrPipe()
	utils.CheckError(err)

	err = cmd.Start()
	if err != nil {
		log.Errorf("%s", err.Error())
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	err = cmd.Wait()
}
