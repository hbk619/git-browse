//go:build linux
// +build linux

package notifications

import (
	"fmt"

	"github.com/hbk619/gh-peruse/internal/requests"
)

func Notify(message string, command requests.CommandLine) error {
	result, err := command.Run("notify-send", []string{message})
	if err != nil {
		return err
	}

	if result != "" {
		return fmt.Errorf("failed to notify %w", err)
	}

	return nil
}
