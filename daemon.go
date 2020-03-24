// +build !windows

package docker

import (
	"io/ioutil"
	"os"
)

const dockerExe = "/usr/local/bin/docker"
const dockerdExe = "/usr/local/bin/dockerd"
const dockerHome = "/root/.docker/"

func (p Plugin) startDaemon() {
	cmd := commandDaemon(p.Daemon)
	if p.Daemon.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
	}
	go func() {
		trace(cmd)
		cmd.Run()
	}()
}
