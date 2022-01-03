package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-github/v34/github"
	"github.com/ktrysmt/go-bitbucket"
	gitlab "github.com/xanzy/go-gitlab"
)

var (
	GitHubClient    *github.Client
	GitLabClient    *gitlab.Client
	BitbucketClient *bitbucket.Client
	mux             *http.ServeMux
	server          *httptest.Server
)

func setup() {
	os.Setenv("GITHUB_TOKEN", "$$$randome")
	os.Setenv("GITLAB_TOKEN", "$$$randome")
	os.Setenv("BITBUCKET_USERNAME", "bbuser")
	os.Setenv("BITBUCKET_PASSWORD", "$$$randomp")

	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	base, _ := url.Parse(server.URL)

	// Add a trailing slash because GitHub SDK expects it
	u, err := url.Parse("/")
	if err != nil {
		log.Fatal(err)
	}
	url := base.ResolveReference(u)

	// github client configured to use test server
	GitHubClient = github.NewClient(nil)
	GitHubClient.BaseURL = url

	// github client configured to use test server
	GitLabClient, err = gitlab.NewClient("", gitlab.WithBaseURL(url.String()))

	BitbucketClient = bitbucket.NewBasicAuth(os.Getenv("BITBUCKET_USERNAME"), os.Getenv("BITBUCKET_USERNAME"))
	BitbucketClient.SetApiBaseURL(url.String())
}

func teardown() {
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITLAB_TOKEN")
	os.Unsetenv("BITBUCKET_USERNAME")
	os.Unsetenv("BITBUCKET_PASSWORD")
	server.Close()
}

func TestGetPublicGitHubRepositories(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"full_name": "test/r1", "id":1, "ssh_url": "https://github.com/u/r1", "name": "r1", "private": false, "fork": false}]`)
	})

	repos, err := getRepositories(GitHubClient, "github", "all", "", "", false)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var expected []*Repository
	expected = append(expected, &Repository{Namespace: "test", CloneURL: "https://github.com/u/r1", Name: "r1", Private: false})
	if !reflect.DeepEqual(repos, expected) {
		t.Errorf("Expected %+v, Got %+v", expected, repos)
	}
}

func TestGetPrivateGitHubRepositories(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"full_name": "test/r1", "id":1, "ssh_url": "https://github.com/u/r1", "name": "r1", "private": true, "fork": false}]`)
	})

	repos, err := getRepositories(GitHubClient, "github", "all", "", "", false)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var expected []*Repository
	expected = append(expected, &Repository{Namespace: "test", CloneURL: "https://github.com/u/r1", Name: "r1", Private: true})
	if !reflect.DeepEqual(repos, expected) {
		t.Errorf("Expected %+v, Got %+v", expected, repos)
	}
}

func TestGetStarredGitHubRepositories(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/starred", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"repo":{"full_name": "test/r1", "id":1, "ssh_url": "https://github.com/u/r1", "name": "r1", "private": true, "fork": false}}]`)
	})

	repos, err := getRepositories(GitHubClient, "github", "starred", "", "", false)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var expected []*Repository
	expected = append(expected, &Repository{Namespace: "test", CloneURL: "https://github.com/u/r1", Name: "r1", Private: true})
	if !reflect.DeepEqual(repos, expected) {
		t.Errorf("Expected %+v, Got %+v", expected, repos)
	}
}

func TestGetGitLabRepositories(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"path_with_namespace": "test/r1", "id":1, "ssh_url_to_repo": "https://gitlab.com/u/r1", "name": "r1"}]`)
	})

	repos, err := getRepositories(GitLabClient, "gitlab", "internal", "", "", false)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var expected []*Repository
	expected = append(expected, &Repository{Namespace: "test", CloneURL: "https://gitlab.com/u/r1", Name: "r1"})
	if !reflect.DeepEqual(repos, expected) {
		for i := 0; i < len(repos); i++ {
			t.Errorf("Expected %+v, Got %+v", expected[i], repos[i])
		}
	}
}

func TestGetStarredGitLabRepositories(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%#v\n", r.URL.Query())
		if len(r.URL.Query().Get("starred")) != 0 {
			fmt.Fprint(w, `[{"path_with_namespace": "test/starred-repo-r1", "id":1, "ssh_url_to_repo": "https://gitlab.com/u/r1", "name": "starred-repo-r1"}]`)
			return
		}
		fmt.Fprintf(w, `[]`)
	})

	repos, err := getRepositories(GitLabClient, "gitlab", "", "", "starred", false)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var expected []*Repository
	expected = append(expected, &Repository{Namespace: "test", CloneURL: "https://gitlab.com/u/r1", Name: "starred-repo-r1"})

	if !reflect.DeepEqual(repos, expected) {
		if len(repos) != len(expected) {
			t.Fatalf("Expected: %#v, Got: %v", expected, repos)
		}
		for i := 0; i < len(expected); i++ {
			t.Errorf("Expected %+v, Got %+v", expected[i], repos[i])
		}
	}
}

func TestGetBitbucketRepositories(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/workspaces", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"pagelen": 10, "page": 1, "size": 1, "values": [{"slug": "abc"}]}`)
	})

	mux.HandleFunc("/repositories/abc", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"pagelen": 10, "page": 1, "size": 1, "values": [{"full_name":"abc/def", "slug":"def", "is_private":true, "links":{"clone":[{"name":"https", "href":"https://bbuser@bitbucket.org/abc/def.git"}, {"name":"ssh", "href":"git@bitbucket.org:abc/def.git"}]}}]}`)
	})

	repos, err := getRepositories(BitbucketClient, "bitbucket", "", "", "", false)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var expected []*Repository
	expected = append(expected, &Repository{Namespace: "abc", CloneURL: "git@bitbucket.org:abc/def.git", Name: "def", Private: true})
	if !reflect.DeepEqual(repos, expected) {
		for i := 0; i < len(repos); i++ {
			t.Errorf("Expected %+v, Got %+v", expected[i], repos[i])
		}
	}
}
