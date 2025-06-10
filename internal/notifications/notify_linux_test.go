//go:build linux

package notifications

import (
	"testing"

	"github.com/golang/mock/gomock"
	mock_requests "github.com/hbk619/gh-peruse/internal/requests/mocks"
	"github.com/stretchr/testify/suite"
)

type NotifyLinuxSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockCommandLine *mock_requests.MockCommandLine
}

func (suite *NotifyLinuxSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockCommandLine = mock_requests.NewMockCommandLine(suite.ctrl)
}

func (suite *NotifyLinuxSuite) TestNotify_sends_msg() {
	suite.mockCommandLine.EXPECT().Run("notify-send", []string{"test"}).Return("", nil)
	Notify("test", suite.mockCommandLine)
}

func TestNotifyLinuxSuiteSuite(t *testing.T) {
	suite.Run(t, new(NotifyLinuxSuite))
}
