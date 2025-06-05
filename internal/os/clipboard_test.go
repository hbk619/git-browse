package os

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hbk619/gh-peruse/internal/requests"
	"github.com/stretchr/testify/suite"
)

type ClipboardSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	originalWriteTo func(contents string, command requests.CommandLine) error
	clipboard       *Clipboard
}

func (suite *ClipboardSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.originalWriteTo = WriteTo
	suite.clipboard = NewClipboard()
}

func (suite *ClipboardSuite) AfterTest(string, string) {
	WriteTo = suite.originalWriteTo
}

func (suite *ClipboardSuite) TestWrite_calls_write_to() {
	actualContents := ""
	WriteTo = func(contents string, command requests.CommandLine) error {
		actualContents = contents
		return nil
	}
	err := suite.clipboard.Write("test")
	suite.NoError(err)
	suite.Equal("test", actualContents)
}

func (suite *ClipboardSuite) TestWrite_returns_error() {
	expectedErr := errors.New("oops")
	WriteTo = func(contents string, command requests.CommandLine) error {
		return expectedErr
	}
	actualErr := suite.clipboard.Write("test")
	suite.ErrorIs(actualErr, expectedErr)
}

func TestClipboardSuite(t *testing.T) {
	suite.Run(t, new(ClipboardSuite))
}
