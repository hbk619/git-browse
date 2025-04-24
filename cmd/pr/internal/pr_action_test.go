package internal

import (
	"errors"
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
	suite.mockPrClient.EXPECT().GetPRDetails(repo, false).Return(&git.PR{
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
	suite.mockPrClient.EXPECT().GetPRDetails(repo, false).Return(&git.PR{
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
	suite.mockPrClient.EXPECT().GetPRDetails(repo, false).Return(&git.PR{
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

func (suite *PRActionTestSuite) TestInit_verbose_prints_state_and_gets_commit_comments() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: {CommentCount: 4}}}).Return(nil)
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(repo, true).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State: git.State{
			MergeStatus:    "mergable",
			ConflictStatus: "no conflicts",
			Reviews:        map[string][]string{"APPROVED": {"Peach", "Goomba"}},
			Statuses: []git.Status{{
				Name:       "ci",
				Conclusion: "SUCCESS",
			}, {
				Name:       "build",
				Conclusion: "FAILURE",
			}},
		},
		Title: "A spiffing PR",
	}, nil)
	suite.mockPrClient.EXPECT().GetCommitComments(repo.Owner, repo.Name, 2).Return([]git.Comment{
		{Body: "Comment 3", Author: git.Author{Login: "Mario"}},
		{Body: "Comment 4", Author: git.Author{Login: "Toad"}},
	}, nil)

	suite.mockOutput.EXPECT().Print("mergable")
	suite.mockOutput.EXPECT().Print("no conflicts")
	suite.mockOutput.EXPECT().Print("APPROVED Peach Goomba")
	suite.mockOutput.EXPECT().Print("Check ci SUCCESS")
	suite.mockOutput.EXPECT().Print("Check build FAILURE")
	suite.mockOutput.EXPECT().Print("New comments ahead!")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init(2, true)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_no_new_comments_since_last_view() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}).Return(nil)
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(repo, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init(2, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_err_getting_repo() {
	expectedErr := errors.New("failed to get repo")
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(nil, expectedErr)

	err := suite.prAction.Init(2, false)
	suite.ErrorIs(err, expectedErr)
}

func (suite *PRActionTestSuite) TestInit_err_getting_comments() {
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	expectedErr := errors.New("failed to get comments")
	suite.mockPrClient.EXPECT().GetPRDetails(repo, false).Return(nil, expectedErr)

	err := suite.prAction.Init(2, false)
	suite.ErrorIs(err, expectedErr)
}

func (suite *PRActionTestSuite) TestInit_err_getting_commit_comments() {
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	expectedErr := errors.New("failed to get commit comments")
	suite.mockPrClient.EXPECT().GetPRDetails(repo, true).Return(&git.PR{
		Comments: []git.Comment{},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockPrClient.EXPECT().GetCommitComments(repo.Owner, repo.Name, 2).Return(nil, expectedErr)

	err := suite.prAction.Init(2, true)
	suite.ErrorIs(err, expectedErr)
}

func (suite *PRActionTestSuite) TestInit_err_loading_history() {
	expectedErr := errors.New("no permission to read file")
	suite.mockHistory.EXPECT().Load().Return(history.History{}, expectedErr)
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(repo, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("Warning failed to load comments to history: no permission to read file")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init(2, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_err_saving_history() {
	expectedErr := errors.New("no permission to write file")
	prHistory := history.PR{CommentCount: 2}
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{2: {CommentCount: 1}}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: prHistory}}).Return(expectedErr)
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(repo, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("Warning failed to save comments to history: no permission to write file")
	suite.mockOutput.EXPECT().Print("New comments ahead!")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init(2, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestPrint_prints_threads() {
	suite.prAction.Results = []git.Comment{{
		Body: "Comment 1",
		Thread: git.Thread{
			IsResolved: true,
		},
		Author: git.Author{
			Login: "Mario",
		},
	}}

	suite.mockOutput.EXPECT().Print("This comment is resolved")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	suite.prAction.Print()
}

func (suite *PRActionTestSuite) TestPrint_prints_outdated() {
	suite.prAction.Results = []git.Comment{{
		Body:     "Comment 1",
		Outdated: true,
		Author: git.Author{
			Login: "Mario",
		},
	}}

	suite.mockOutput.EXPECT().Print("This comment is outdated")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	suite.prAction.Print()
}

func TestPrActionSuite(t *testing.T) {
	suite.Run(t, new(PRActionTestSuite))
}
