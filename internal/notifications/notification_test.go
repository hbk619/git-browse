package notifications

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hbk619/gh-peruse/internal/requests"
	"github.com/stretchr/testify/suite"
)

type NotifierSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	originalNotify func(contents string, command requests.CommandLine) error
	Notifier       *Notifier
}

func (suite *NotifierSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.originalNotify = Notify
	suite.Notifier = NewNotifier()
}

func (suite *NotifierSuite) AfterTest(string, string) {
	Notify = suite.originalNotify
}

func (suite *NotifierSuite) TestWrite_calls_write_to() {
	actualContents := ""
	Notify = func(contents string, command requests.CommandLine) error {
		actualContents = contents
		return nil
	}
	err := suite.Notifier.Println("test")
	suite.NoError(err)
	suite.Equal("test", actualContents)
}

func (suite *NotifierSuite) TestWrite_returns_error() {
	expectedErr := errors.New("oops")
	Notify = func(contents string, command requests.CommandLine) error {
		return expectedErr
	}
	actualErr := suite.Notifier.Println("test")
	suite.ErrorIs(actualErr, expectedErr)
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierSuite))
}
