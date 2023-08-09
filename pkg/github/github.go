package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const RepoURL = "https://api.github.com/search/repositories"

type TrendingSearchResult struct {
	TotalCount int
	Items      []*RepoTrending
}

type RepoTrending struct {
	Full_name         string
	Html_url          string //`json:"html_url"`
	Description       string
	Created_at        time.Time //`json:"created_at"`
	Updated_at        time.Time //`json:"updated_at"`
	Pushed_at         time.Time //`json:"pushed_at"`
	Size              int
	Language          string
	Stargazers_count  int
	Forks_count       int
	Archived          bool
	Open_issues_count int
	Topics            []string
}

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
