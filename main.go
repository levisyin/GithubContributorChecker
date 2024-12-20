package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v56/github"
)

var (
	proxy    = flag.String("proxy", "", "proxy setting")
	repos    = flag.String("repo", "golang/go", "checking repos, split using comma")
	interval = flag.Int64("interval", 1000, "unit ms")
	anon     = flag.Bool("anon", false, "include anonymous contributors in results or not")
	token    = flag.String("token", "", "github token")
	useCache = flag.Bool("useCache", false, "whether to use local cache")
	debug    = flag.Bool("debug", false, "whether to print debug log")
)

func main() {
	flag.Parse()
	repositories := strings.Split(*repos, ",")
	httpCli := http.DefaultClient
	if *proxy != "" {
		var proxyUrl *url.URL
		proxyUrl, _ = url.Parse(*proxy)
		httpCli = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	}
	githubCli := github.NewClient(httpCli).WithAuthToken(*token)
	fmt.Println("gonna check repos:", repositories)
	for _, repository := range repositories {
		splitAry := strings.Split(repository, "/")
		owner := splitAry[0]
		repo := splitAry[1]
		contributors := make([]*github.Contributor, 0)
		if *useCache && localCacheExists(owner, repo) {
			slog.With("owner", owner, "repo", repo).Info("using local cache data")
			err := loadLocalCache(owner, repo, &contributors)
			if err != nil {
				slog.With("err", err, "owner", owner, "repo", repo).Error("load cache from local error")
				continue
			}
		} else {
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
					slog.With("err", err, "owner", owner, "repo", repo).Error("ListContributors error")
					time.Sleep(3 * time.Second)
					continue
				}
				slog.With("page", page, "owner", owner, "repo", repo).Info("ListContributors success")
				contributors = append(contributors, contributorsTmp...)
				if len(contributorsTmp) < 100 {
					break
				}
				time.Sleep(time.Duration(*interval) * time.Millisecond)
			}
			fmt.Printf("Found %d contributors in repo %s\n", len(contributors), repository)
			storeToLocal(owner, repo, contributors)
		}

		for i, contributor := range contributors {
			if *contributor.Type == "Anonymous" {
				slog.With("owner", owner, "repo", repo, "commits", *contributor.Contributions, "name", *contributor.Name, "email", *contributor.Email).Info("anonymous user")
				continue
			}
			// {"login":"rigelrozanski","id":20132176,"node_id":"MDQ6VXNlcjIwMTMyMTc2","avatar_url":"https://avatars.githubusercontent.com/u/20132176?v=4","gravatar_id":"","url":"https://api.github.com/users/rigelrozanski","html_url":"https://github.com/rigelrozanski","followers_url":"https://api.github.com/users/rigelrozanski/followers","following_url":"https://api.github.com/users/rigelrozanski/following{/other_user}","gists_url":"https://api.github.com/users/rigelrozanski/gists{/gist_id}","starred_url":"https://api.github.com/users/rigelrozanski/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/rigelrozanski/subscriptions","organizations_url":"https://api.github.com/users/rigelrozanski/orgs","repos_url":"https://api.github.com/users/rigelrozanski/repos","events_url":"https://api.github.com/users/rigelrozanski/events{/privacy}","received_events_url":"https://api.github.com/users/rigelrozanski/received_events","type":"User","site_admin":false,"contributions":922}
			fmt.Printf("%s: order: %d, user: %s[%d], commits: %d, home: %s\n", repository, i+1, *contributor.Login, *contributor.ID, *contributor.Contributions, *contributor.HTMLURL)
			if strings.Contains(*contributor.Login, "[bot]") {
				continue
			}
			rsp, err := httpCli.Get("https://github.com/" + *contributor.Login)
			if err != nil {
				slog.With("err", err, "owner", owner, "repo", repo).Error("get user info error")
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

func getLocalCacheFile(owner, repo string) string {
	return fmt.Sprintf("%s__%s", owner, repo) + ".json"
}

func loadLocalCache(owner, repo string, recv any) error {
	file, err := os.Open(getLocalCacheFile(owner, repo))
	if err != nil {
		slog.With("err", err, "owner", owner, "repo", repo).Error("open file error")
		return err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		slog.With("err", err, "owner", owner, "repo", repo).Error("io read error")
		return err
	}
	if err := json.Unmarshal(byteValue, recv); err != nil {
		slog.With("err", err, "owner", owner, "repo", repo).Error("unmarshal error")
		return err
	}
	return nil
}

func storeToLocal(owner, repo string, data any) error {
	file, err := os.Create(getLocalCacheFile(owner, repo))
	if err != nil {
		slog.With("err", err, "owner", owner, "repo", repo).Error("create file error")
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err := encoder.Encode(data); err != nil {
		slog.With("err", err, "owner", owner, "repo", repo).Error("encode data to file error")
		return err
	}
	slog.With("owner", owner, "repo", repo).Info("store local cache success")
	return nil
}

// localCacheExists check if local cache exists
func localCacheExists(owner, repo string) bool {
	_, err := os.Stat(getLocalCacheFile(owner, repo))
	return !errors.Is(err, os.ErrNotExist)
}

func initLog() {
	defaultLevel := slog.LevelInfo
	if *debug {
		defaultLevel = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(defaultLevel)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: defaultLevel,
	}))
	slog.SetDefault(logger)
}
