package internal

import (
	"github.com/golang/mock/gomock"
	mock_filesystem "github.com/hbk619/git-browse/internal/filesystem/mocks"
	"github.com/hbk619/git-browse/internal/git"
	mock_github "github.com/hbk619/git-browse/internal/github/mocks"
	"github.com/hbk619/git-browse/internal/history"
	mock_history "github.com/hbk619/git-browse/internal/history/mocks"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PRActionTestSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	prAction     *PRAction
	mockHistory  *mock_history.MockStorage
	mockOutput   *mock_filesystem.MockOutput
	mockPrClient *mock_github.MockPullRequestClient
}

func (suite *PRActionTestSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockOutput = mock_filesystem.NewMockOutput(suite.ctrl)
	suite.mockHistory = mock_history.NewMockStorage(suite.ctrl)
	suite.mockPrClient = mock_github.NewMockPullRequestClient(suite.ctrl)
	suite.prAction = NewPRAction(suite.mockPrClient, suite.mockHistory, suite.mockOutput)
}

func (suite *PRActionTestSuite) TestInit_no_comments() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: make(map[int]history.PR)}, nil)
	prHistory := history.PR{CommentCount: 0}
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: prHistory}}).Return(nil)
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetMainPRDetails(2, false).Return(&git.PR{
		Comments: []git.Comment{},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	err := suite.prAction.Init(2, false)
	suite.ErrorContains(err, "no comments found")
}

func (suite *PRActionTestSuite) TestInit_new_comments_never_viewed_pr() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}).Return(nil)
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetMainPRDetails(2, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("New comments ahead!")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init(2, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_new_comments_since_last_view() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{2: {CommentCount: 1}}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}).Return(nil)
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetMainPRDetails(2, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("New comments ahead!")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init(2, false)
	suite.NoError(err)
}

func TestPrActionSuite(t *testing.T) {
	suite.Run(t, new(PRActionTestSuite))
}
