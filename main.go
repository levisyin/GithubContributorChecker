package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v56/github"
)

var (
	proxy    = flag.String("proxy", "", "proxy setting")
	repos    = flag.String("repo", "golang/go", "checking repos, split using a comma")
	interval = flag.Int64("interval", 1000, "unit ms")
	anon     = flag.Bool("anon", false, "include anonymous contributors in results or not")
	token    = flag.String("token", "", "github token, couldn't be empty")
)

func main() {
	flag.Parse()
  if *token == "" {
    panic("github token shouln't be empty")
  }
	repositories := strings.Split(*repos, ",")
	httpCli := http.DefaultClient
	if *proxy != "" {
		var proxyURL *url.URL
		proxyURL, _ = url.Parse(*proxy)
		httpCli = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	}
	githubCli := github.NewClient(httpCli).WithAuthToken(*token)
	for _, repository := range repositories {
		splitAry := strings.Split(repository, "/")
		owner := splitAry[0]
		repo := splitAry[1]
		contributors := make([]*github.Contributor, 0)
		page := 0
		size := 100
		for {
			page = page + 1
			contributorsTmp, _, err := githubCli.Repositories.ListContributors(context.Background(), owner, repo, &github.ListContributorsOptions{
				Anon: fmt.Sprintf("%v", *anon),
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: size,
				},
			})
			if err != nil {
				fmt.Println("ListContributors error", err)
				continue
			}
			contributors = append(contributors, contributorsTmp...)
			if len(contributorsTmp) < 100 {
				break
			}
			time.Sleep(time.Duration(*interval) * time.Millisecond)
		}
		fmt.Printf("Found %d contributors in repo %s\n", len(contributors), repository)
		for i, contributor := range contributors {
			// {"login":"rigelrozanski","id":20132176,"node_id":"MDQ6VXNlcjIwMTMyMTc2","avatar_url":"https://avatars.githubusercontent.com/u/20132176?v=4","gravatar_id":"","url":"https://api.github.com/users/rigelrozanski","html_url":"https://github.com/rigelrozanski","followers_url":"https://api.github.com/users/rigelrozanski/followers","following_url":"https://api.github.com/users/rigelrozanski/following{/other_user}","gists_url":"https://api.github.com/users/rigelrozanski/gists{/gist_id}","starred_url":"https://api.github.com/users/rigelrozanski/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/rigelrozanski/subscriptions","organizations_url":"https://api.github.com/users/rigelrozanski/orgs","repos_url":"https://api.github.com/users/rigelrozanski/repos","events_url":"https://api.github.com/users/rigelrozanski/events{/privacy}","received_events_url":"https://api.github.com/users/rigelrozanski/received_events","type":"User","site_admin":false,"contributions":922}
			fmt.Printf("%s: order: %d, user: %s[%d], commits: %d, home: %s\n", repository, i+1, *contributor.Login, *contributor.ID, *contributor.Contributions, *contributor.HTMLURL)
			if strings.Contains(*contributor.Login, "[bot]") {
				continue
			}
			rsp, err := httpCli.Get("https://github.com/" + *contributor.Login)
			if err != nil {
				fmt.Println("ListContributors error", err)
				time.Sleep(time.Duration(*interval) * time.Millisecond)
				continue
			}
			if rsp.StatusCode == http.StatusNotFound {
				fmt.Println("-------------------------------------------------------------------------------------------------------")
				fmt.Printf("-----------------------------------FOUND A USER: %s---------------------------------\n", *contributor.Login)
				fmt.Println("-------------------------------------------------------------------------------------------------------")
			}
			time.Sleep(time.Duration(*interval) * time.Millisecond)
		}
	}
}
