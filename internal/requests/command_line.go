package requests

import (
	"bytes"
	"os/exec"
	"strings"
)

type (
	CommandLine interface {
		Run(command string) (string, error)
	}

	Bash struct{}
)

func NewBash() *Bash {
	return &Bash{}
}

func (bash *Bash) Run(command string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
