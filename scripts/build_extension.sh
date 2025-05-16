#!/usr/bin/zsh

go build
GOOS=windows GOARCH=amd64 go build -o gh-peruse-windows-amd64.exe
GOOS=linux GOARCH=amd64 go build -o gh-peruse-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o gh-peruse-darwin-amd64