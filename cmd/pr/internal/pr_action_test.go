package internal

import (
	"errors"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/golang/mock/gomock"
	mock_filesystem "github.com/hbk619/gh-peruse/internal/filesystem/mocks"
	"github.com/hbk619/gh-peruse/internal/git"
	"github.com/hbk619/gh-peruse/internal/github"
	mock_github "github.com/hbk619/gh-peruse/internal/github/mocks"
	"github.com/hbk619/gh-peruse/internal/history"
	mock_history "github.com/hbk619/gh-peruse/internal/history/mocks"
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
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(suite.prAction.Repo, false).Return(&git.PR{
		Comments: []git.Comment{},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	err := suite.prAction.Init([]string{"2"}, false)
	suite.ErrorContains(err, "no comments found")
}

func (suite *PRActionTestSuite) TestInit_gets_pr_number() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: make(map[int]history.PR)}, nil)
	prHistory := history.PR{CommentCount: 0}
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: prHistory}}).Return(nil)
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().DetectCurrentPR(suite.prAction.Repo).Return(2, nil)
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(suite.prAction.Repo, false).Return(&git.PR{
		Comments: []git.Comment{},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	err := suite.prAction.Init([]string{}, false)
	suite.ErrorContains(err, "no comments found")
}

func (suite *PRActionTestSuite) TestInit_error_when_getting_pr_number() {
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	expected := errors.New("mocked error")
	suite.mockPrClient.EXPECT().DetectCurrentPR(suite.prAction.Repo).Return(0, expected)
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)

	err := suite.prAction.Init([]string{}, false)
	suite.ErrorContains(err, "mocked error")
}

func (suite *PRActionTestSuite) TestInit_invalid_pr_number() {
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)

	err := suite.prAction.Init([]string{"asdasd"}, false)
	suite.ErrorContains(err, "provide a valid PR number")
}

func (suite *PRActionTestSuite) TestInit_new_comments_never_viewed_pr() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}).Return(nil)
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(suite.prAction.Repo, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("New comments ahead!")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init([]string{"2"}, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_new_comments_since_last_view() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{2: {CommentCount: 1}}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}).Return(nil)
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(suite.prAction.Repo, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("New comments ahead!")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init([]string{"2"}, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_verbose_prints_state() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}).Return(nil)
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(suite.prAction.Repo, true).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}},
			{Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
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

	suite.mockOutput.EXPECT().Print("mergable")
	suite.mockOutput.EXPECT().Print("no conflicts")
	suite.mockOutput.EXPECT().Print("APPROVED Peach Goomba")
	suite.mockOutput.EXPECT().Print("Check ci SUCCESS")
	suite.mockOutput.EXPECT().Print("Check build FAILURE")
	suite.mockOutput.EXPECT().Print("New comments ahead!")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init([]string{"2"}, true)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_no_new_comments_since_last_view() {
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: {CommentCount: 2}}}).Return(nil)
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(suite.prAction.Repo, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init([]string{"2"}, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_err_getting_repo() {
	expectedErr := errors.New("failed to get repo")
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repository.Repository{}, expectedErr)

	err := suite.prAction.Init([]string{"2"}, false)
	suite.ErrorIs(err, expectedErr)
}

func (suite *PRActionTestSuite) TestInit_err_getting_comments() {
	githubRepo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	repo := &git.Repo{
		Owner:    "Bowser",
		Name:     "castle",
		PRNumber: 2,
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(githubRepo, nil)
	expectedErr := errors.New("failed to get comments")
	suite.mockPrClient.EXPECT().GetPRDetails(repo, false).Return(nil, expectedErr)

	err := suite.prAction.Init([]string{"2"}, false)
	suite.ErrorIs(err, expectedErr)
}

func (suite *PRActionTestSuite) TestInit_err_loading_history() {
	expectedErr := errors.New("no permission to read file")
	suite.mockHistory.EXPECT().Load().Return(history.History{}, expectedErr)
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(suite.prAction.Repo, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("Warning failed to load comments to history: no permission to read file")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init([]string{"2"}, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestInit_err_saving_history() {
	expectedErr := errors.New("no permission to write file")
	prHistory := history.PR{CommentCount: 2}
	suite.mockHistory.EXPECT().Load().Return(history.History{Prs: map[int]history.PR{2: {CommentCount: 1}}}, nil)
	suite.mockHistory.EXPECT().Save(history.History{Prs: map[int]history.PR{2: prHistory}}).Return(expectedErr)
	repo := repository.Repository{
		Owner: "Bowser",
		Name:  "castle",
	}
	suite.mockPrClient.EXPECT().GetRepoDetails().Return(repo, nil)
	suite.mockPrClient.EXPECT().GetPRDetails(suite.prAction.Repo, false).Return(&git.PR{
		Comments: []git.Comment{{Body: "Comment 1", Author: git.Author{Login: "Mario"}}, {Body: "Comment 2", Author: git.Author{Login: "Peach"}}},
		State:    git.State{},
		Title:    "A spiffing PR",
	}, nil)

	suite.mockOutput.EXPECT().Print("Warning failed to save comments to history: no permission to write file")
	suite.mockOutput.EXPECT().Print("New comments ahead!")
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	err := suite.prAction.Init([]string{"2"}, false)
	suite.NoError(err)
}

func (suite *PRActionTestSuite) TestPrint_prints_resolved_threads() {
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
	suite.prAction.Print()
}

func (suite *PRActionTestSuite) TestPrint_prints_file_info_with_path() {
	suite.prAction.Results = []git.Comment{{
		Body: "Comment 1",
		Author: git.Author{
			Login: "Mario",
		},
		File: git.File{FullPath: github.MainThread},
	}, {
		Body: "Comment 2",
		Author: git.Author{
			Login: "Luigi",
		},
		File: git.File{FullPath: "README.md:28", Path: "/", Line: 28, LineContents: "whhhaaayy", FileName: "README.md"},
	}}

	suite.prAction.LastFullPath = github.MainThread
	suite.prAction.Index = 1
	suite.mockOutput.EXPECT().Print("README.md")
	suite.mockOutput.EXPECT().Print("/")
	suite.mockOutput.EXPECT().Print("28")
	suite.mockOutput.EXPECT().Print("whhhaaayy")
	suite.mockOutput.EXPECT().Print("Luigi")
	suite.mockOutput.EXPECT().Print("Comment 2")
	suite.prAction.Print()
}

func (suite *PRActionTestSuite) TestPrint_prints_file_info_without_path() {
	suite.prAction.Results = []git.Comment{{
		Body: "Comment 1",
		Author: git.Author{
			Login: "Mario",
		},
		File: git.File{FullPath: github.MainThread, FileName: github.MainThread},
	}, {
		Body: "Comment 2",
		Author: git.Author{
			Login: "Luigi",
		},
		File: git.File{FullPath: "README.md:8", Path: "/", Line: 28, LineContents: "whhhaaayy"},
	}}

	suite.prAction.LastFullPath = "README.md:8"
	suite.prAction.Index = 0
	suite.mockOutput.EXPECT().Print(github.MainThread)
	suite.mockOutput.EXPECT().Print("Mario")
	suite.mockOutput.EXPECT().Print("Comment 1")
	suite.prAction.Print()
}

func (suite *PRActionTestSuite) TestReply() {
	comments := []git.Comment{{
		Id: "awdasdadad",
		Author: git.Author{
			Login: "Bowser",
		},
		Body: "Rraaawwww",
	}, {
		Id: "23213213",
		Author: git.Author{
			Login: "Peach",
		},
		Body:  "Great start",
		State: "COMMENTED",
	}, {
		Id: "lkmoimiom",
		Author: git.Author{
			Login: "Yoshi",
		},
		Body: "Yum!",
	},
	}
	repo := &git.Repo{
		Owner:    "mario",
		Name:     "kart",
		PRNumber: 2,
	}
	suite.prAction.Repo = repo
	suite.prAction.Results = comments
	suite.mockPrClient.EXPECT().Reply("ta", &comments[0], suite.prAction.Id).Return(nil)
	suite.mockOutput.EXPECT().Print("Posted comment")
	suite.prAction.Reply("ta")
}

func (suite *PRActionTestSuite) TestReply_middle_comment() {
	comments := []git.Comment{{
		Id: "awdasdadad",
		Author: git.Author{
			Login: "Bowser",
		},
		Body: "Rraaawwww",
	}, {
		Id: "23213213",
		Author: git.Author{
			Login: "Peach",
		},
		Body:  "Great start",
		State: "COMMENTED",
	}, {
		Id: "lkmoimiom",
		Author: git.Author{
			Login: "Yoshi",
		},
		Body: "Yum!",
	},
	}
	repo := &git.Repo{
		Owner:    "mario",
		Name:     "kart",
		PRNumber: 2,
	}
	suite.prAction.Repo = repo
	suite.prAction.Results = comments
	suite.prAction.Index = 1
	suite.mockPrClient.EXPECT().Reply("ta", &comments[1], suite.prAction.Id).Return(nil)
	suite.mockOutput.EXPECT().Print("Posted comment")
	suite.prAction.Reply("ta")
}

func (suite *PRActionTestSuite) TestReply_print_error() {
	comments := []git.Comment{{
		Id: "awdasdadad",
		Author: git.Author{
			Login: "Bowser",
		},
		Body: "Rraaawwww",
	}, {
		Id: "23213213",
		Author: git.Author{
			Login: "Peach",
		},
		Body:  "Great start",
		State: "COMMENTED",
	}, {
		Id: "lkmoimiom",
		Author: git.Author{
			Login: "Yoshi",
		},
		Body: "Yum!",
	},
	}
	repo := &git.Repo{
		Owner:    "mario",
		Name:     "kart",
		PRNumber: 2,
	}
	suite.prAction.Repo = repo
	suite.prAction.Results = comments
	suite.prAction.Index = 1
	suite.mockPrClient.EXPECT().Reply("ta", &comments[1], suite.prAction.Id).Return(errors.New("some error"))
	suite.mockOutput.EXPECT().Print("Warning failed to comment: some error")
	suite.prAction.Reply("ta")
}

func (suite *PRActionTestSuite) TestResolve() {
	comments := []git.Comment{{
		Id: "awdasdadad",
		Author: git.Author{
			Login: "Bowser",
		},
		Body: "Rraaawwww",
		Thread: git.Thread{
			IsResolved: false,
			ID:         "1223333",
		},
	}, {
		Id: "23213213",
		Author: git.Author{
			Login: "Peach",
		},
		Body:  "Great start",
		State: "COMMENTED",
	}, {
		Id: "lkmoimiom",
		Author: git.Author{
			Login: "Yoshi",
		},
		Body: "Yum!",
	},
	}
	suite.prAction.Results = comments
	suite.mockPrClient.EXPECT().Resolve(&comments[0]).Return(nil)
	suite.mockOutput.EXPECT().Print("Conversation resolved")
	suite.prAction.Resolve()
}

func (suite *PRActionTestSuite) TestResolve_middle_comment() {
	comments := []git.Comment{{
		Id: "awdasdadad",
		Author: git.Author{
			Login: "Bowser",
		},
		Body: "Rraaawwww",
		Thread: git.Thread{
			IsResolved: false,
			ID:         "1223333",
		},
	}, {
		Id: "23213213",
		Author: git.Author{
			Login: "Peach",
		},
		Body:  "Great start",
		State: "COMMENTED",
		Thread: git.Thread{
			IsResolved: false,
			ID:         "4343434",
		},
	}, {
		Id: "lkmoimiom",
		Author: git.Author{
			Login: "Yoshi",
		},
		Body: "Yum!",
	},
	}
	suite.prAction.Results = comments
	suite.prAction.Index = 1
	suite.mockPrClient.EXPECT().Resolve(&comments[1]).Return(nil)
	suite.mockOutput.EXPECT().Print("Conversation resolved")
	suite.prAction.Resolve()
}

func (suite *PRActionTestSuite) TestResolve_error() {
	comments := []git.Comment{{
		Id: "awdasdadad",
		Author: git.Author{
			Login: "Bowser",
		},
		Body: "Rraaawwww",
		Thread: git.Thread{
			IsResolved: false,
			ID:         "1223333",
		},
	}, {
		Id: "23213213",
		Author: git.Author{
			Login: "Peach",
		},
		Body:  "Great start",
		State: "COMMENTED",
		Thread: git.Thread{
			IsResolved: false,
			ID:         "4343434",
		},
	}, {
		Id: "lkmoimiom",
		Author: git.Author{
			Login: "Yoshi",
		},
		Body: "Yum!",
	},
	}
	suite.prAction.Results = comments
	suite.prAction.Index = 2
	suite.mockPrClient.EXPECT().Resolve(&comments[2]).Return(errors.New("some error"))
	suite.mockOutput.EXPECT().Print("Warning failed to resolve thread: some error")
	suite.prAction.Resolve()
}

func TestPrActionSuite(t *testing.T) {
	suite.Run(t, new(PRActionTestSuite))
}
