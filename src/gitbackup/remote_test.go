package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
)

var (
	client *github.Client
	mux    *http.ServeMux
	server *httptest.Server
)

func setup() {
	os.Setenv("GITHUB_TOKEN", "$$$randome")
	os.Setenv("GITLAB_TOKEN", "$$$randome")

	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// github client configured to use test server
	client = github.NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
	client.UploadURL = url
}

func teardown() {
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITLAB_TOKEN")
	server.Close()
}

func TestNewClient(t *testing.T) {
	setup()
	defer teardown()

	client := NewClient("github")
	// Type assertion
	client = client.(*github.Client)

	client = NewClient("gitlab")
	// Type assertion
	client = client.(*gitlab.Client)

	client = NewClient("notyetsupported")
	if client != nil {
		t.Errorf("Expected nil")
	}

}

func TestGetRepositories(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"id":1, "git_url": "git://github.com/u/r1", "name": "r1"}]`)
	})

	repos, err := getRepositories(client, "github", "", "all", "")
	if err != nil {
		t.Fatal("%v", err)
	}
	var expected []*Repository
	expected = append(expected, &Repository{GitURL: "git://github.com/u/r1", Name: "r1"})
	if !reflect.DeepEqual(repos, expected) {
		t.Errorf("Expected %+v, Got %+v", expected, repos)
	}
}
