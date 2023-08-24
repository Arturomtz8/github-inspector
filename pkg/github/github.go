package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// TrendingSearchResult struct holds an slice of trending repositories on GitHub and its count.
type TrendingSearchResult struct {
	TotalCount int
	Items      []*RepoTrending
}

// RepoTrending is the treding repository reprentation.
type RepoTrending struct {
	FullName    string `json:"full_name"`
	HtmlURL     string `json:"html_url"`
	Description string
	Owner       struct {
		Login string
	}
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        time.Time `json:"pushed_at"`
	Size            int
	Language        string
	StargazersCount int `json:"stargazers_count"`
	ForksCount      int `json:"forks_count"`
	Archived        bool
	OpenIssuesCount int `json:"open_issues_count"`
	Topics          []string
}

// SearchGithubTrending function returns a list treding repositores on GitHub.
// The function returns this list in the form of a pointer to a TrendingSearchResult.
// If some thing wrong happens, it returns an error.
func SearchGithubTrending(term, APIurl string) (*TrendingSearchResult, error) {
	// in case receiving more values, consider changing to slice term string[]
	// q := url.QueryEscape(strings.Join(terms, " "))
	term = url.QueryEscape(term)
	// https://api.github.com/search/issues?q=stress+test+label:bug+language:python+state:closed&per_page=100
	resp, err := http.Get(APIurl + "?q=stars:<=500+archived:false+language:" + term + "&per_page=5&sort=stars&order=desc")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search query failed: %s", resp.Status)

	}

	var result TrendingSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	result.TotalCount = len(result.Items)
	return &result, nil
}
