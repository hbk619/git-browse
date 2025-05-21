package new_comments

import (
	"errors"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/golang/mock/gomock"
	mock_filesystem "github.com/hbk619/gh-peruse/internal/filesystem/mocks"
	"github.com/hbk619/gh-peruse/internal/git"
	mock_github "github.com/hbk619/gh-peruse/internal/github/mocks"
	"github.com/hbk619/gh-peruse/internal/history"
	mock_history "github.com/hbk619/gh-peruse/internal/history/mocks"
	"github.com/stretchr/testify/suite"
)

type CheckNewComments struct {
	suite.Suite
	ctrl         *gomock.Controller
	mockHistory  *mock_history.MockStorage
	mockOutput   *mock_filesystem.MockOutput
	mockPrClient *mock_github.MockPullRequestClient
	repo         repository.Repository
	gitRepo      *git.Repo
}

func (suite *CheckNewComments) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockOutput = mock_filesystem.NewMockOutput(suite.ctrl)
	suite.mockHistory = mock_history.NewMockStorage(suite.ctrl)
	suite.mockPrClient = mock_github.NewMockPullRequestClient(suite.ctrl)
	suite.repo = repository.Repository{
		Owner: "luigi",
		Name:  "mansion",
	}
	suite.gitRepo = &git.Repo{
		Owner: "luigi",
		Name:  "mansion",
	}
}

func (suite *CheckNewComments) TestCheckForNewComments_finds_comments() {
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(suite.repo, nil)
	suite.mockPrClient.EXPECT().GetCommentCountForOwnedPRs(suite.gitRepo).
		Return(map[int]int{
			2: 3,
			1: 7,
			3: 1,
		}, nil)
	suite.mockHistory.EXPECT().Load().Return(history.History{
		Prs: map[int]history.PR{
			2: history.PR{
				CommentCount: 2,
			},
			3: history.PR{
				CommentCount: 1,
			},
		},
	}, nil)
	suite.mockOutput.EXPECT().Println("Pull request 1 has new comments")
	suite.mockOutput.EXPECT().Println("Pull request 2 has new comments")

	err := CheckForNewComments(suite.mockPrClient, suite.mockHistory, suite.mockOutput)
	suite.NoError(err)
}

func (suite *CheckNewComments) TestCheckForNewComments_returns_errors_from_fetching_comment_count() {
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(suite.repo, nil)
	suite.mockPrClient.EXPECT().GetCommentCountForOwnedPRs(suite.gitRepo).Return(nil, errors.New("failed to get comments"))

	err := CheckForNewComments(suite.mockPrClient, suite.mockHistory, suite.mockOutput)
	suite.ErrorContains(err, "failed to get comments")
}

func (suite *CheckNewComments) TestCheckForNewComments_returns_errors_from_fetching_history() {
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(suite.repo, nil)
	suite.mockPrClient.EXPECT().GetCommentCountForOwnedPRs(suite.gitRepo).
		Return(map[int]int{
			2: 3,
			1: 7,
			3: 1,
		}, nil)
	suite.mockHistory.EXPECT().Load().Return(history.History{}, errors.New("failed to get history"))

	err := CheckForNewComments(suite.mockPrClient, suite.mockHistory, suite.mockOutput)
	suite.ErrorContains(err, "failed to get history")
}
func (suite *CheckNewComments) TestCheckForNewComments_returns_errors_from_output() {
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(suite.repo, nil)
	suite.mockPrClient.EXPECT().GetCommentCountForOwnedPRs(suite.gitRepo).
		Return(map[int]int{
			2: 3,
			1: 7,
			3: 1,
		}, nil)
	suite.mockHistory.EXPECT().Load().Return(history.History{
		Prs: map[int]history.PR{
			2: history.PR{
				CommentCount: 2,
			},
			3: history.PR{
				CommentCount: 1,
			},
		},
	}, nil)
	suite.mockOutput.EXPECT().Println("Pull request 1 has new comments").Return(errors.New("failed to print"))
	suite.mockOutput.EXPECT().Println("Pull request 2 has new comments")
	err := CheckForNewComments(suite.mockPrClient, suite.mockHistory, suite.mockOutput)
	suite.ErrorContains(err, "failed to print")
}
func (suite *CheckNewComments) TestCheckForNewComments_returns_errors_from_repo() {
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repository.Repository{}, errors.New("bad repo"))

	err := CheckForNewComments(suite.mockPrClient, suite.mockHistory, suite.mockOutput)
	suite.ErrorContains(err, "bad repo")
}

func TestCheckNewCommentsSuite(t *testing.T) {
	suite.Run(t, new(CheckNewComments))
}
