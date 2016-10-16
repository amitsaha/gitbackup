package main

import (
    "fmt"
    "github.com/google/go-github/github"
)

func main() {
    client := github.NewClient(nil)
    user := "amitsaha"
    // Pagination?
    opt := &github.RepositoryListOptions{Type: "owner", Sort: "updated", Direction: "desc"}
    repos, _, err := client.Repositories.List(user, opt)
    if err != nil {
            fmt.Println(err)
    }
    for _, repo := range repos {
        // Better way to print dereferenced pointers?
        fmt.Printf("Name: %v GitURL: %v\n", *repo.Name, *repo.GitURL)
        //fmt.Println(repo.String())
    }
}
