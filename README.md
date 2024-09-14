# GithubContributorChecker
A tool that check whether if contributor changed his name 

## Requirements

- You should set the GitHub token by the `-token` flag. See [Github API](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repository-contributors) for detail.

## QuickStart

```shell
go run .
```

## Examples

1. Check contributors of `ethereum/go-ethereum` repository, running `go run . --repos=ethereum/go-ethereum`
