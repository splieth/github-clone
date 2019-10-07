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
)

type Repository struct {
	Name string
	Url  string
}

type Config struct {
	Token        string
	Organization string
	Destination  string
}

func (c *Config) validate() error {
	if c.Token == "" {
		return fmt.Errorf("token must be provided")
	}
	if c.Organization == "" {
		return fmt.Errorf("orga must be provided")
	}
	if c.Destination == "" {
		return fmt.Errorf("dest must be provided")
	}
	return nil
}

func main() {
	config := Config{}
	flag.StringVar(&config.Token, "token", "", "Your GitHub token")
	flag.StringVar(&config.Organization, "orga", "", "Name of the GitHub organization")
	flag.StringVar(&config.Destination, "dest", "", "Destination folder")
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

	repos, err := getAllRepos(client, ctx, config.Organization)
	if err != nil {
		fmt.Printf("Error: ", err)
		os.Exit(1)
	}

	fmt.Println("Found ", len(repos), " repositories. Updating now...")
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
	var cmd = exec.Command(
		"cd", repo.Name, "&&",
		"git", "stash", "&&",
		"git", "pull", "&&",
		"git", "stash", "pop", )
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("git pull failed for ", repo.Name)
		os.Exit(1)
	}
}

func getAllRepos(client *github.Client, ctx context.Context, organization string) ([]Repository, error) {
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
		repos = append(repos, Repository{
			Name: *repo.Name,
			Url:  *repo.SSHURL,
		})
	}
	return repos, nil
}
