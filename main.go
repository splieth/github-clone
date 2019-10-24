package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"os"
	"path"
	"strings"
)

type Repository struct {
	Name string
	Url  string
}

type Config struct {
	Token        string
	Organization string
	Destination  string
	HostReplace  string
	FailOnError  bool
}

func checkError(err error, exit bool) {
	if err != nil {
		fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
		if exit {
			os.Exit(1)
		}
	}
}

func (c *Config) validate() error {
	if c.Token == "" {
		return fmt.Errorf("token must be provided")
	}
	if c.Organization == "" {
		return fmt.Errorf("org must be provided")
	}
	if c.Destination == "" {
		return fmt.Errorf("dest must be provided")
	}
	if c.Destination == "" {
		return fmt.Errorf("fail-on-error must be provided")
	}
	return nil
}

func main() {
	config := Config{}
	flag.StringVar(&config.Token, "token", "", "Your GitHub token")
	flag.StringVar(&config.Organization, "org", "", "Name of the GitHub organization")
	flag.StringVar(&config.Destination, "destination", "", "Destination folder")
	flag.StringVar(&config.HostReplace, "host", "", "Replacement for github.com in SSH URL, e.g. if you use multiple SSH keys for GitHub")
	flag.BoolVar(&config.FailOnError, "fail-on-error", false, "Fail if git clone/git pull fail. Continues by default")
	flag.Parse()
	err := config.validate()
	checkError(err, true)

	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token}),
	))

	repos, err := getAllRepos(client, ctx, config.Organization, config.HostReplace)
	checkError(err, true)

	fmt.Println("Found", len(repos), "repositories. Updating now...")
	for _, repo := range repos {
		updateRepo(repo, config)
	}
}

func updateRepo(repo Repository, config Config) {
	pathToRepo := path.Join(config.Destination, repo.Name)
	if _, err := os.Stat(pathToRepo); os.IsNotExist(err) {
		fmt.Println(repo.Name, " not present, cloning...")
		clone(repo, pathToRepo, config.FailOnError)
	} else {
		fmt.Println(repo.Name, " already there, updating...")
		pull(repo, pathToRepo, config.FailOnError)
	}
}

func clone(repo Repository, path string, fail bool) {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      repo.Url,
		Progress: os.Stdout,
	})
	checkError(err, fail)
}

func pull(repo Repository, path string, fail bool) {
	r, err := git.PlainOpen(path)
	checkError(err, fail)
	w, err := r.Worktree()
	checkError(err, fail)
	err = w.Pull(&git.PullOptions{})
	ref, err := r.Head()
	checkError(err, fail)
	commit, err := r.CommitObject(ref.Hash())
	checkError(err, fail)

	fmt.Println("Latest commit for", repo.Name, "->", strings.TrimSuffix(commit.Message, "\n"), " (by", commit.Author.Name, ")")
}

func getAllRepos(client *github.Client, ctx context.Context, organization, hostReplace string) ([]Repository, error) {
	var result []*github.Repository

	listByOrgOptions := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, organization, listByOrgOptions)
		if err != nil {
			return nil, err
		}
		result = append(result, repos...)
		nextPage := resp.NextPage
		if nextPage == 0 {
			break
		}
		listByOrgOptions.ListOptions.Page = nextPage
	}
	var repos []Repository
	for _, repo := range result {
		url := *repo.SSHURL
		if hostReplace != "" {
			url = strings.ReplaceAll(url, "github.com", hostReplace)
		}
		repos = append(repos, Repository{
			Name: *repo.Name,
			Url:  url,
		})
	}
	return repos, nil
}
