//go:build darwin
// +build darwin

package os

import (
	"github.com/hbk619/gh-peruse/internal/requests"
)

var WriteTo = func (contents string, command requests.CommandLine) error {
	_, err := command.RunWithInput("pbcopy", []string{}, contents)

	return err
}
