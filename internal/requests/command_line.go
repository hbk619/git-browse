package requests

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type (
	CommandLine interface {
		Run(executable string, args []string) (string, error)
		RunWithInput(executable string, args []string, input string) (string, error)
	}

	CommandRunner struct{}
)

func NewCommandRunner() *CommandRunner {
	return &CommandRunner{}
}

func (runner *CommandRunner) Run(executable string, args []string) (string, error) {
	executablePath, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	var errorPipe bytes.Buffer

	cmd := exec.Command(executablePath, args...)
	cmd.Stdout = &out
	cmd.Stderr = &errorPipe

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %w", errorPipe.String(), err)
	}

	return strings.TrimSpace(out.String()), nil
}

func (runner *CommandRunner) RunWithInput(executable string, args []string, input string) (string, error) {
	executablePath, err := exec.LookPath(executable)
	if err != nil {
		return "", fmt.Errorf("failed to find executable %w", err)
	}
	var out bytes.Buffer
	var errorPipe bytes.Buffer
	cmd := exec.Command(executablePath, args...)
	cmd.Stdout = &out
	cmd.Stderr = &errorPipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin pipe %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("%s %w", errorPipe.String(), err)
	}

	stdin.Write(fmt.Appendf(nil, "%s\n", input))
	if err := stdin.Close(); err != nil {
		return "", fmt.Errorf("%s %w", errorPipe.String(), err)
	}

	return strings.TrimSpace(out.String()), nil
}
