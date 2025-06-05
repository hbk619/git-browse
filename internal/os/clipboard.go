package os

import (
	"fmt"

	"github.com/hbk619/gh-peruse/internal/requests"
)

type (
	Clipboard struct {
	}
	Clippy interface {
		Write(contents string) error
	}
)

func NewClipboard() *Clipboard {
	return &Clipboard{}
}

func (Clipboard *Clipboard) Write(contents string) error {
	command := requests.NewCommandRunner()
	err := WriteTo(contents, command)
	if err != nil {
		return fmt.Errorf("error copying %w", err)
	}
	return nil
}
