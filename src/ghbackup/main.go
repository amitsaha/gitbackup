package main

import (
    "fmt"
    "github.com/google/go-github/github"
)

func main() {
    client := github.NewClient(nil)
    user := "amitsaha"
    opt := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc"}
    repos, _, err := client.Repositories.List(user, opt)
    if err != nil {
            fmt.Println(err)
    }
    for _, repo := range repos {
        fmt.Printf("Name: %s GitURL: %s\n", repo.Name, repo.GitURL)
        //fmt.Println(repo.String())
    }
}
