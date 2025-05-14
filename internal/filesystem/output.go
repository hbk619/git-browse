package filesystem

import "fmt"

type (
	Output interface {
		Println(text string)
		Print(text string)
	}

	StdOut struct{}
)

func NewStdOut() *StdOut {
	return &StdOut{}
}

func (stdOut *StdOut) Println(text string) {
	fmt.Println(text)
}

func (stdOut *StdOut) Print(text string) {
	fmt.Print(text)
}
