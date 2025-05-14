package internal

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type InteractiveTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	interactive *Interactive
}

func (suite *InteractiveTestSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.interactive = &Interactive{}
}

func (suite *InteractiveTestSuite) TestNext_moves_on_when_less_than_max_index() {
	called := false
	print := func() {
		called = true
	}

	suite.interactive.MaxIndex = 2
	suite.interactive.Index = 1

	suite.interactive.Next(print)

	suite.Equal(true, called)
	suite.Equal(2, suite.interactive.Index)
}

func (suite *InteractiveTestSuite) TestNext_does_not_move_on_when_equal_to_max_index() {
	called := false
	print := func() {
		called = true
	}

	suite.interactive.MaxIndex = 2
	suite.interactive.Index = 2

	suite.interactive.Next(print)

	suite.Equal(false, called)
	suite.Equal(2, suite.interactive.Index)
}

func (suite *InteractiveTestSuite) TestPrevious_moves_on_when_not_on_first_item() {
	called := false
	print := func() {
		called = true
	}

	suite.interactive.MaxIndex = 2
	suite.interactive.Index = 1

	suite.interactive.Previous(print)

	suite.Equal(true, called)
	suite.Equal(0, suite.interactive.Index)
}

func (suite *InteractiveTestSuite) TestPrevious_does_not_move_on_when_on_first_item() {
	called := false
	print := func() {
		called = true
	}

	suite.interactive.MaxIndex = 2
	suite.interactive.Index = 0

	suite.interactive.Previous(print)

	suite.Equal(false, called)
	suite.Equal(0, suite.interactive.Index)
}

func (suite *InteractiveTestSuite) TestRepeat_calls_print() {
	called := false
	print := func() {
		called = true
	}

	suite.interactive.MaxIndex = 2
	suite.interactive.Index = 1

	suite.interactive.Repeat(print)

	suite.Equal(true, called)
	suite.Equal(1, suite.interactive.Index)
}

func TestInteractiveSuite(t *testing.T) {
	suite.Run(t, new(InteractiveTestSuite))
}
