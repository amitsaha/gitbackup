package main

import (
    "fmt"
    "os/exec"
    "github.com/google/go-github/github"
)

func main() {
    client := github.NewClient(nil)
    // Make these configurable
    user := "amitsaha"
    repoType := "all"
    opt := &github.RepositoryListOptions{Type: repoType, Sort: "created", Direction: "desc"}
    for {
        repos, resp, err := client.Repositories.List(user, opt)
        if err != nil {
                fmt.Println(err)
        } else {
            for _, repo := range repos {
                // Better way to print dereferenced pointers?
                fmt.Printf("Name: %v GitURL: %v\n", *repo.Name, *repo.GitURL)
                cmd := exec.Command("git", "clone", *repo.GitURL)
                err := cmd.Run()
                if err != nil {
                    fmt.Printf("Error cloning %v\n", *repo.Name)
                }
            }
        }

        // Learn about GitHub's pagination here:
        // https://developer.github.com/v3/
        if resp.NextPage == 0 {
            break
        }
        opt.ListOptions.Page = resp.NextPage
    }
}
