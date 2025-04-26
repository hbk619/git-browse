package requests

import (
	"bytes"
	"fmt"
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
	var errorPipe bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &out
	cmd.Stderr = &errorPipe
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %w", errorPipe.String(), err)
	}
	return strings.TrimSpace(out.String()), nil
}
