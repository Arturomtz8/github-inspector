package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	RepoURL string = "https://api.github.com/search/repositories"
	// Standard mode: github.com/trending
	modeRepositories = "repositories"
	// Base URL for the github website
	defaultBaseURL = "https://github.com"
	// Relative URL for trending repositories
	urlTrendingPath = "/trending"
	// Relative URL for trending developers
	urlDevelopersPath = "/developers"
	// Developers mode: github.com/trending/developers
	modeDevelopers = "developers"
)

// Owner struct is the author of the repo.
type Owner struct {
	Login string
}

// TrendingSearchResult struct holds an slice of trending repositories on GitHub and its count.
type TrendingSearchResult struct {
	TotalCount int
	Items      []*RepoTrending
}

// RepoTrending is the treding repository reprentation.
type RepoTrending struct {
	FullName        string `json:"full_name"`
	HtmlURL         string `json:"html_url"`
	Description     string
	Owner           Owner
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
	resp, err := http.Get(APIurl + "?q=stars:1000..15000+archived:false+language:" + term + "&per_page=5&sort=stars&order=desc")
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

// GetProjects provides a slice of Projects filtered by the given time and language.
func GetTrendingRepos(time, language, trendingURL string) (*TrendingSearchResult, error) {
	var repos TrendingSearchResult

	// Generate the correct URL to call
	u, err := generateURL(modeRepositories, time, language)
	if err != nil {
		return nil, err
	}

	// Receive document
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Query our information
	doc.Find(".Box article.Box-row").Each(func(_ int, s *goquery.Selection) {
		// Collect project information
		name := getProjectName(s.Find("h2 a").Text())

		// Split name (like "andygrunwald/go-trending") into owner ("andygrunwald") and repository name ("go-trending"")
		splittedName := strings.SplitAfterN(name, "/", 2)

		owner := splittedName[0][:len(splittedName[0])-1]
		owner = strings.TrimSpace(owner)
		repositoryName := strings.TrimSpace(splittedName[1])

		// Overwrite name to be 100% sure it contains no space between owner and repo name
		name = fmt.Sprintf("%s/%s", owner, repositoryName)

		address, exists := s.Find("h2 a").First().Attr("href")
		projectURL := appendBaseHostToPath(defaultBaseURL, address, exists)

		description := s.Find("p").Text()
		description = strings.TrimSpace(description)

		language := s.Find("span[itemprop=programmingLanguage]").Eq(0).Text()
		language = strings.TrimSpace(language)

		starsString := s.Find("div a[href$=\"/stargazers\"]").Text()
		starsString = strings.TrimSpace(starsString)

		// Replace english thousand separator ","
		starsString = strings.Replace(starsString, ",", "", 1)
		stars, err := strconv.Atoi(starsString)
		if err != nil {
			stars = 0
		}

		p := &RepoTrending{
			FullName: name,
			Owner: Owner{
				Login: owner,
			},
			HtmlURL:         projectURL.String(),
			Description:     description,
			Language:        language,
			StargazersCount: stars,
		}

		repos.Items = append(repos.Items, p)
	})

	return &repos, nil
}

// appendBaseHostToPath will add the base host to a relative url urlStr.
// A urlStr like "/trending" will be returned as https://github.com/trending
func appendBaseHostToPath(baseurl, urlStr string, exists bool) *url.URL {
	baseURL, _ := url.Parse(baseurl)

	if !exists {
		return nil
	}

	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}

	return baseURL.ResolveReference(rel)
}

// generateURL will generate the correct URL to call the github site.
//
// Depending on mode, time and language it will set the correct pathes and query parameters.
func generateURL(mode, time, language string) (*url.URL, error) {
	urlStr := urlTrendingPath
	if mode == modeDevelopers {
		urlStr += urlDevelopersPath
	}

	u := appendBaseHostToPath(defaultBaseURL, urlStr, true)
	q := u.Query()
	if len(time) > 0 {
		q.Set("since", time)
	}

	if len(language) > 0 {
		q.Set("l", language)
	}

	u.RawQuery = q.Encode()

	return u, nil
}

// getProjectName will return the project name in format owner/repository
func getProjectName(name string) string {
	trimmedNameParts := []string{}

	nameParts := strings.Split(name, "\n")
	for _, part := range nameParts {
		trimmedNameParts = append(trimmedNameParts, strings.TrimSpace(part))
	}

	return strings.Join(trimmedNameParts, "")
}
