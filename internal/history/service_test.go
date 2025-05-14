package history

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	mock_filesystem "github.com/hbk619/gh-peruse/internal/filesystem/mocks"
	"github.com/stretchr/testify/suite"
)

type HistoryServiceTestSuite struct {
	suite.Suite
	mockFilesystem *mock_filesystem.MockFS
	ctrl           *gomock.Controller
	historyService *Service
}

func (suite *HistoryServiceTestSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockFilesystem = mock_filesystem.NewMockFS(suite.ctrl)
	suite.historyService = &Service{
		configPath: "config/path",
		fs:         suite.mockFilesystem,
	}
}

func (suite *HistoryServiceTestSuite) TestLoad_returns_default_when_no_history() {
	notFound := fs.ErrNotExist
	suite.mockFilesystem.EXPECT().ReadFile("config/path").Return(nil, notFound)
	history, err := suite.historyService.Load()
	suite.NoError(err)
	suite.Equal(History{
		Prs: make(map[int]PR),
	}, history)
}

func (suite *HistoryServiceTestSuite) TestLoad_returns_history_when_exists() {
	expectedHistory := History{
		Prs: map[int]PR{
			2: {
				CommentCount: 4,
			},
			3: {
				CommentCount: 8,
			},
		},
	}
	suite.mockFilesystem.EXPECT().ReadFile("config/path").Return([]byte(`{"Prs":{"2":{"CommentCount": 4},"3":{"CommentCount":8}}}`), nil)
	history, err := suite.historyService.Load()
	suite.NoError(err)
	suite.Equal(expectedHistory, history)
}

func (suite *HistoryServiceTestSuite) TestLoad_returns_default_and_err_when_reading_fails() {
	streamClosed := fs.ErrClosed
	suite.mockFilesystem.EXPECT().ReadFile("config/path").Return(nil, streamClosed)
	history, err := suite.historyService.Load()
	suite.ErrorIs(err, streamClosed)
	suite.Equal(History{}, history)
}

func (suite *HistoryServiceTestSuite) TestLoad_returns_default_and_err_when_json_invalid() {
	suite.mockFilesystem.EXPECT().ReadFile("config/path").Return([]byte(`{Prs":{"2":{"CommentCount": 4},"3":{"CommentCount":8}}}`), nil)
	history, err := suite.historyService.Load()
	suite.ErrorContains(err, "invalid character")
	suite.Equal(History{}, history)
}

func (suite *HistoryServiceTestSuite) TestSave_saves_history() {
	history := History{
		Prs: map[int]PR{
			2: {
				CommentCount: 4,
			},
			3: {
				CommentCount: 8,
			},
		},
	}
	suite.mockFilesystem.EXPECT().SaveFile("config/path", []byte(`{"Prs":{"2":{"CommentCount":4},"3":{"CommentCount":8}}}`))
	err := suite.historyService.Save(history)
	suite.NoError(err)
}

func (suite *HistoryServiceTestSuite) TestSave_returns_error_if_save_fails() {
	history := History{
		Prs: map[int]PR{
			2: {
				CommentCount: 4,
			},
			3: {
				CommentCount: 8,
			},
		},
	}
	expectedError := errors.New("uh oh")
	suite.mockFilesystem.EXPECT().SaveFile("config/path", []byte(`{"Prs":{"2":{"CommentCount":4},"3":{"CommentCount":8}}}`)).
		Return(expectedError)
	err := suite.historyService.Save(history)
	suite.ErrorIs(err, expectedError)
}

func (suite *HistoryServiceTestSuite) TestNewHistoryService() {
	expectedService := &Service{
		configPath: path.Join("base/path", ".config", "gh-peruse-history.json"),
		fs:         suite.mockFilesystem,
	}
	suite.mockFilesystem.EXPECT().MkdirAll(path.Join("base/path", ".config"), os.ModeDir).Return(nil)
	service, err := NewHistoryService("base/path", suite.mockFilesystem)
	suite.NoError(err)
	suite.Equal(service, expectedService)
}

func (suite *HistoryServiceTestSuite) TestNewHistoryService_returns_err_if_config_dir_cannot_be_made() {
	expectedError := errors.New("bad fs!")
	suite.mockFilesystem.EXPECT().MkdirAll(path.Join("base/path", ".config"), os.ModeDir).Return(expectedError)
	service, err := NewHistoryService("base/path", suite.mockFilesystem)
	suite.ErrorIs(err, expectedError)
	suite.Nil(service)
}

func TestHistoryServiceSuite(t *testing.T) {
	suite.Run(t, new(HistoryServiceTestSuite))
}
