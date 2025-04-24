package github

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/hbk619/git-browse/internal/requests"
	mock_requests "github.com/hbk619/git-browse/internal/requests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type APITestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockCommandLine *mock_requests.MockCommandLine
	ghApi           *GHApi
}

func (suite *APITestSuite) BeforeTest(string, string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockCommandLine = mock_requests.NewMockCommandLine(suite.ctrl)
	httpClient := requests.NewAuthorisedHTTPClient(&requests.AuthorisedHTTPClientOptions{AuthToken: "Authy boy"})
	suite.ghApi = &GHApi{
		httpClient:        httpClient,
		commandLineClient: suite.mockCommandLine,
	}
}

func (suite *APITestSuite) TestGHApi_LoadGitHubAPIJSON() {
	suite.mockCommandLine.EXPECT().Run("run me").Return("[{}, {}, {}]", nil)
	got, err := suite.ghApi.LoadGitHubAPIJSON("run me")
	suite.NoError(err)
	suite.Equal("[{}, {}, {}]", string(got))
}

func (suite *APITestSuite) TestGHApi_LoadGitHubAPIJSON_command_err() {
	expected := errors.New("oops")
	suite.mockCommandLine.EXPECT().Run("run me").Return("", expected)
	got, err := suite.ghApi.LoadGitHubAPIJSON("run me")
	suite.Error(err)
	suite.Nil(got)
}

func (suite *APITestSuite) TestGHApi_LoadGitHubAPIJSON_gh_pr_not_found() {
	suite.mockCommandLine.EXPECT().Run("run me").Return(`[{"message":"Not found"}]`, nil)
	got, err := suite.ghApi.LoadGitHubAPIJSON("run me")
	suite.ErrorContains(err, "pull request not found")
	suite.Nil(got)
}

func (suite *APITestSuite) TestGHApi_LoadGitHubAPIJSON_gh_commit_not_found() {
	suite.mockCommandLine.EXPECT().Run("run me").Return(`[{"message":"No commit found"}]`, nil)
	got, err := suite.ghApi.LoadGitHubAPIJSON("run me")
	suite.ErrorContains(err, "commit not found")
	suite.Nil(got)
}

func (suite *APITestSuite) TestGHApi_LoadGitHubAPIJSON_gh_other_error() {
	suite.mockCommandLine.EXPECT().Run("run me").Return(`[{"message":"bad things happened"}]`, nil)
	got, err := suite.ghApi.LoadGitHubAPIJSON("run me")
	suite.ErrorContains(err, "bad things happened")
	suite.Nil(got)
}

func mockGraphQL(t *testing.T, responseWriter http.ResponseWriter, request *http.Request, body string) {
	assert.Equal(t, request.Method, http.MethodPost)
	assert.Equal(t, "Bearer Authy boy", request.Header.Get("Authorization"))

	responseWriter.Header().Set("Content-Type", "application/json")

	_, err := responseWriter.Write([]byte(body))

	if err != nil {
		t.Errorf("expected err to be nil got %v", err)
	}
}

func (suite *APITestSuite) TestGHApi_LoadGitHubGraphQLJSON() {
	variables := map[string]interface{}{
		"test": 1,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/graphql":
			mockGraphQL(suite.T(), w, r, `{"Data":{}}`)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))
	defer svr.Close()
	suite.T().Setenv("GITHUB_GRAPHQL", svr.URL+"/graphql")

	got, err := suite.ghApi.LoadGitHubGraphQLJSON("query()", variables)
	suite.NoError(err)
	suite.Equal(`{"Data":{}}`, string(got))
}

func (suite *APITestSuite) TestGHApi_LoadGitHubGraphQLJSON_has_error() {
	variables := map[string]interface{}{
		"test": 1,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/graphql":
			mockGraphQL(suite.T(), w, r, `{"Errors":[{"message":"not found"},{"message":"bad things happened"}]}`)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))
	defer svr.Close()
	suite.T().Setenv("GITHUB_GRAPHQL", svr.URL+"/graphql")

	got, err := suite.ghApi.LoadGitHubGraphQLJSON("query()", variables)
	suite.ErrorContains(err, "not found bad things happened")
	suite.Equal("", string(got))
}

func (suite *APITestSuite) TestGHApi_LoadGitHubGraphQLJSON_url_not_found() {
	variables := map[string]interface{}{
		"test": 1,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/graphql":
			http.NotFoundHandler().ServeHTTP(w, r)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))
	defer svr.Close()
	url := svr.URL + "/graphql"
	suite.T().Setenv("GITHUB_GRAPHQL", url)

	got, err := suite.ghApi.LoadGitHubGraphQLJSON("query()", variables)
	suite.ErrorContains(err, url+" not found")
	suite.Equal("", string(got))
}

func (suite *APITestSuite) TestGHApi_LoadGitHubGraphQLJSON_url_forbidden() {
	variables := map[string]interface{}{
		"test": 1,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/graphql":
			http.Error(w, "bad", http.StatusForbidden)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))
	defer svr.Close()
	url := svr.URL + "/graphql"
	suite.T().Setenv("GITHUB_GRAPHQL", url)

	got, err := suite.ghApi.LoadGitHubGraphQLJSON("query()", variables)
	suite.ErrorContains(err, "not allowed to view url "+url)
	suite.Equal("", string(got))
}

func (suite *APITestSuite) TestGHApi_LoadGitHubGraphQLJSON_url_unauthorised() {
	variables := map[string]interface{}{
		"test": 1,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/graphql":
			http.Error(w, "bad", http.StatusUnauthorized)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))
	defer svr.Close()
	url := svr.URL + "/graphql"
	suite.T().Setenv("GITHUB_GRAPHQL", url)

	got, err := suite.ghApi.LoadGitHubGraphQLJSON("query()", variables)
	suite.ErrorContains(err, "bad auth token")
	suite.Equal("", string(got))
}

func (suite *APITestSuite) TestGHApi_LoadGitHubGraphQLJSON_url_bad_request() {
	variables := map[string]interface{}{
		"test": 1,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/graphql":
			http.Error(w, "bad", http.StatusBadRequest)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))
	defer svr.Close()
	url := svr.URL + "/graphql"
	suite.T().Setenv("GITHUB_GRAPHQL", url)

	got, err := suite.ghApi.LoadGitHubGraphQLJSON("query()", variables)
	suite.ErrorContains(err, "something bad happened 400 Bad Request")
	suite.Equal("", string(got))
}

func (suite *APITestSuite) TestGHApi_LoadGitHubGraphQLJSON_url_teapot() {
	variables := map[string]interface{}{
		"test": 1,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/graphql":
			http.Error(w, "bad", http.StatusTeapot)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))
	defer svr.Close()
	url := svr.URL + "/graphql"
	suite.T().Setenv("GITHUB_GRAPHQL", url)

	got, err := suite.ghApi.LoadGitHubGraphQLJSON("query()", variables)
	suite.ErrorContains(err, "something bad happened 418 I'm a teapot")
	suite.Equal("", string(got))
}

func TestApiSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
