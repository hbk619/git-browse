package internal

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	mock_filesystem "github.com/hbk619/gh-peruse/internal/filesystem/mocks"
	"github.com/stretchr/testify/suite"
)

type PromptTestSuite struct {
	suite.Suite
	mockOutput *mock_filesystem.MockOutput
	ctrl       *gomock.Controller
	prompt     *Prompter
}

func (suite *PromptTestSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockOutput = mock_filesystem.NewMockOutput(suite.ctrl)
	suite.prompt = &Prompter{
		output: suite.mockOutput,
	}
}

func (suite *PromptTestSuite) TestString_trims_whitespace() {
	suite.mockOutput.EXPECT().Print("enter things: ")
	suite.prompt.input = strings.NewReader("      got it!     \n")
	result := suite.prompt.String("enter things")

	suite.Equal("got it!", result)
}

func (suite *PromptTestSuite) TestString_reads_until_new_line() {
	suite.mockOutput.EXPECT().Print("enter things: ")
	suite.prompt.input = strings.NewReader("got it!\n")
	result := suite.prompt.String("enter things")

	suite.Equal("got it!", result)
}

func TestPromptSuite(t *testing.T) {
	suite.Run(t, new(PromptTestSuite))
}
