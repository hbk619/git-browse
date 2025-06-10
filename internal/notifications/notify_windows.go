//go:build windows
// +build windows

package notifications

import (
	"fmt"

	"github.com/hbk619/gh-peruse/internal/requests"
)

var Notify = func(message string, command requests.CommandLine) error {
	result, err := command.Run("msg", []string{"*", "/TIME:3", message})
	if err != nil {
		return err
	}

	if result != "" {
		return fmt.Errorf("failed to notify %w", err)
	}

	return nil
}
