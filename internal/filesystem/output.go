package filesystem

import "fmt"

type (
	Output interface {
		Println(text string) error
		Print(text string) error
	}

	StdOut struct{}
)

func NewStdOut() *StdOut {
	return &StdOut{}
}

func (stdOut *StdOut) Println(text string) error {
	fmt.Println(text)
	return nil
}

func (stdOut *StdOut) Print(text string) error {
	fmt.Print(text)
	return nil
}
