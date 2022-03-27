package utils

import (
	"fmt"
	utilexec "k8s.io/utils/exec"
	"os"
	"strings"
)

type Exec interface {
	Command(cmd string, args ...string) utilexec.Cmd
}

type executor struct {
	exec utilexec.Interface
}

func NewExecutor() Exec {
	return &executor{
		exec: utilexec.New(),
	}
}

func (e *executor) Command(cmd string, args ...string) utilexec.Cmd {
	fullArgs := append([]string{cmd}, args...)
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(fullArgs, " "))

	return e.exec.Command(cmd, args...)
}
