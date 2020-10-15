// +build windows

package docker

const dockerExe = "C:\\bin\\docker.exe"
const dockerHome = "C:\\ProgramData\\docker\\"

func (p Plugin) startDaemon() {
	// this is a no-op on windows
}
