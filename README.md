# GithubContributorChecker
A tool that checks the contributors of given repositories

## Requirements

- You should set the GitHub token by the `-token` flag. See [Github API](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repository-contributors) for detail.

## Features

- [x] Support scanning multiple GitHub repositories
- [X] Support HTTP proxy settings
- [X] Support scanning frequency control
- [X] Support reading data from local cache
- [ ] Support exporting to file
- [ ] Support multiple condition filters

## QuickStart

```shell
go mod init
go mod tidy
go run .
```

## Examples

1. Check contributors of `ethereum/go-ethereum` repository, running `go run . --repos=ethereum/go-ethereum`
