# Git Browse

`gh peruse pr <pr number>`

For full up-to-date flags:

`gh peruse pr -h`

## Without installing Github CLI
`git-pr` uses the Github CLI library and follows the authentication mechanism and configuration options it offers:

```markdown
GitHub API requests will be authenticated using the same mechanism as gh, i.e. using the values of GH_TOKEN and GH_HOST environment variables and falling back to the user's stored OAuth token
```

You can use `git-pr` [without installing the Github CLI](#without-installing-github-cli) or the [install the Github CLI](#github-cli) to use OAuth.

### Personal Access token
Create a [classic personal access token](https://github.com/settings/tokens/new) with the repo scope
and put it in an environment variable called `GH_TOKEN` or `GITHUB_TOKEN`.

If you are using an enterprise server use `GH_ENTERPRISE_TOKEN` or `GITHUB_ENTERPRISE_TOKEN`

### Changing the host
If the host cannot be inferred from the context of a repo (shouldn't happen often), set an environment variable called `GH_HOST`

### Install the extension

#### MacOS and Linux

Download gh-peruse-darwin-amd64 for Mac or gh-peruse-linux-amd64 for linux from
[latest release](https://github.com/hbk619/gh-peruse/releases/latest) 

Copy the binary to somewhere on your path (e.g. `/usr/local/bin`) called `gh-peruse`

#### Windows

Download gh-peruse-windows-amd64.exe for Windows from
[latest release](https://github.com/hbk619/gh-peruse/releases/latest) 

Copy the binary to somewhere on your path and name it `gh-peruse.exe`

### Usage

To view the current PR for your branch:

`git peruse pr`

To view a specific PR:

`git peruse pr <pr number>`

For full up-to-date flags:

`git peruse pr -h`


## Developing

Install [Go](https://go.dev/doc/install)

### Building

`go build -o git-pr cmd/pr/main.go`

This gives a binary in the dir you run the command from called 'git-pr' which you can run with:

`./git-pr 1`

### Tests

Run with `go test ./...`