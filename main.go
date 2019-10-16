package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
	"os/exec"
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
	return nil
}

func main() {
	config := Config{}
	flag.StringVar(&config.Token, "token", "", "Your GitHub token")
	flag.StringVar(&config.Organization, "org", "", "Name of the GitHub organization")
	flag.StringVar(&config.Destination, "destination", "", "Destination folder")
	flag.StringVar(&config.HostReplace, "host", "", "Replacement for github.com in SSH URL, e.g. if you use multiple SSH keys for GitHub")
	flag.Parse()
	err := config.validate()
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token}),
	))

	repos, err := getAllRepos(client, ctx, config.Organization, config.HostReplace)
	if err != nil {
		fmt.Printf("Error: ", err)
		os.Exit(1)
	}

	fmt.Println("Found", len(repos), "repositories. Updating now...")
	for _, repo := range repos {
		updateRepo(repo, config.Destination)
	}
}

func updateRepo(repo Repository, dest string) {
	pathToRepo := path.Join(dest, repo.Name)
	if _, err := os.Stat(pathToRepo); os.IsNotExist(err) {
		fmt.Println(repo.Name, " not present, cloning...")
		clone(repo, pathToRepo)
	} else {
		fmt.Println(repo.Name, " already there, updating...")
		pull(repo, pathToRepo)
	}
}

func clone(repo Repository, dest string) {
	var cmd = exec.Command("git", "clone", repo.Url, dest)
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("git clone failed for ", repo.Name)
		os.Exit(1)
	}
}

func pull(repo Repository, dest string) {
	cmd := exec.Command("git", "pull", "-r", )
	cwd, _ := os.Getwd()
	_ = os.Chdir(dest)
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("git pull failed for ", repo.Name, " with ", cmd.Stderr)
		os.Exit(1)
	}
	_ = os.Chdir(cwd)
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
