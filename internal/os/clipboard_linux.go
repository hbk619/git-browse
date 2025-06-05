//go:build linux
// +build linux

package os

import (
	"errors"
	"os/exec"

	"github.com/hbk619/gh-peruse/internal/requests"
)

var (
	xsel    = "xsel"
	xclip   = "xclip"
	WriteTo = func(contents string, command requests.CommandLine) error {
		executable := xclip
		args := []string{"-in", "-selection", "clipboard"}
		if _, err := exec.LookPath(xclip); err != nil {
			executable = xsel
			args = []string{"--input", "--clipboard"}
			if _, err := exec.LookPath(xsel); err != nil {
				return errors.New("failed to find xclip or xsel, please install with your package manager")
			}
		}

		_, err := command.RunWithInput(executable, args, contents)

		return err
	}
)
