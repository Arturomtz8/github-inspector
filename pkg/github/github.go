package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const RepoURL = "https://api.github.com/search/repositories"

// TrendingSearchResult struct holds an slice of trending repositories on GitHub and its count.
type TrendingSearchResult struct {
	TotalCount int
	Items      []*RepoTrending
}

// RepoTrending is the treding repository reprentation.
type RepoTrending struct {
	Full_name         string
	Html_url          string
	Description       string
	Created_at        time.Time
	Updated_at        time.Time
	Pushed_at         time.Time
	Size              int
	Language          string
	Stargazers_count  int
	Forks_count       int
	Archived          bool
	Open_issues_count int
	Topics            []string
}

// SearchGithubTrending function returns a list treding repositores on GitHub.
// The function returns this list in the form of a pointer to a TrendingSearchResult.
// If some thing wrong happens, it returns an error.
func SearchGithubTrending(term string) (*TrendingSearchResult, error) {
	// in case receiving more values, consider changing to slice term string[]
	// q := url.QueryEscape(strings.Join(terms, " "))
	term = url.QueryEscape(term)
	// https://api.github.com/search/issues?q=stress+test+label:bug+language:python+state:closed&per_page=100
	resp, err := http.Get(RepoURL + "?q=stars:<=500+archived:false+language:" + term + "&per_page=5&sort=stars&order=desc")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("search query failed: %s", resp.Status)

	}

	var result TrendingSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		resp.Body.Close()
		return nil, err
	}

	resp.Body.Close()

	result.TotalCount = len(result.Items)
	return &result, nil
}
