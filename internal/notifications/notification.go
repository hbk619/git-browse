package notifications

import (
	"fmt"

	"github.com/hbk619/gh-peruse/internal/requests"
)

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (notifier *Notifier) Println(message string) error {
	command := requests.NewCommandRunner()
	err := Notify(message, command)
	if err != nil {
		return fmt.Errorf("error notifying %w", err)
	}
	return nil
}

func (notifier *Notifier) Print(message string) error {
	return notifier.Println(message)
}
