# Git Browse

For using git in a more screen reader friendly way and for those who like to go one at a time

## Pre requisites
Install the [Github CLI](https://cli.github.com/)

Run `gh auth login` and approve the app in the browser. 

## Installation

Download and copy the [latest release](https://github.com/hbk619/git-browse/releases/download/v0.0.5/git-pr) to somewhere on your path (e.g. `/usr/local/bin`)

## Usage

`git pr <pr number>`

For full up-to-date flags:

`git pr -h`

## Developing

Install [Go](https://go.dev/doc/install)

### Building

`go build -o git-pr cmd/pr/main.go`

This gives a binary in the dir you run the command from called 'git-pr' which you can run with:

`./git-pr 1`

### Tests

Run with `go test ./...`