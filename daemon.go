// +build !windows

package docker

import (
	"io/ioutil"
	"os"
	"text/template"
)

const (
	daemonConfigTemplate = `{
		{{ if .HTTPProxy || .HTTPSProxy || .NoProxy }}
		"proxies": {
			"default": {
				{{ if .HTTPProxy }}
				"httpProxy": "{{.HTTPProxy}},
				{{ endif }}
				{{ if .HTTPSProxy }}
				"httpsProxy": "{{.HTTPSProxy}},
				{{ endif }}
				{{ if .NoProxy }}
				"noProxy": "{{.NoProxy}}
				{{ endif }}
			}
		}
		{{ endif }}
	}`

	dockerConfigDir = "/root/.docker/"
	dockerExe       = "/usr/local/bin/docker"
	dockerdExe      = "/usr/local/bin/dockerd"
)

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

// Creates the file needed for dind going through a proxy
func (p Plugin) setProxyConfig() {
	dConfTempl, err := template.New("dConfTempl").Parse(daemonConfigTemplate)
	if err != nil {
		panic("internal error (daemon config template invalid)")
	}
	if _, err := os.Stat(dockerConfigDir); err != nil {
		os.Mkdir(dockerConfigDir, 0755)
	}
	f, err := os.Create(dockerConfigDir + "config.json")
	if err != nil {
		panic(dockerConfigDir + "config.json")
	}
	defer f.Close()
	err = dConfTempl.Execute(f, p.Daemon)
	if err != nil {
		panic("Could not template daemon config")
	}
	f.Sync()
}
