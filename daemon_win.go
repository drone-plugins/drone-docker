// +build windows

package docker

const dockerExe = "C:\\bin\\docker.exe"
const dockerdExe = ""

func (p Plugin) startDaemon() {
	// this is a no-op on windows
}
