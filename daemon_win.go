// +build windows

package docker

const dockerExe = "C:\\bin\\docker.exe"
const dockerdExe = ""
const dockerrootconfdir = "C:\\ProgramData\\docker\"

func (p Plugin) startDaemon() {
	// this is a no-op on windows
}
