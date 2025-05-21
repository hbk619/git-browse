//go:build darwin
// +build darwin

package notifications

import (
	"fmt"

	"github.com/hbk619/gh-peruse/internal/requests"
)

func Notify(message string, command requests.CommandLine) error {
	result, err := command.Run("osascript", []string{"-e", fmt.Sprintf("display notification \"%s\"", message)})
	if err != nil {
		return err
	}

	if result != "" {
		return fmt.Errorf("failed to notify %w", err)
	}

	return nil
}
