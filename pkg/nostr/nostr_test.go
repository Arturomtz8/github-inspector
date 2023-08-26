package nostr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Arturomtz8/github-inspector/pkg/github"
)

func TestGetReposContent(t *testing.T) {
	repos := &github.TrendingSearchResult{
		Items: []*github.RepoTrending{
			{
				FullName:        "foo/bar",
				HtmlURL:         "https://github.com/foo/bar",
				Description:     "a good project",
				Language:        "Go",
				StargazersCount: 1000,
				Owner: github.Owner{
					Login: "foo",
				},
			},
		},
	}

	content, err := getReposContent(repos)
	for _, repo := range content {
		t.Log(repo)
	}
	require.NoError(t, err)
	require.NotEmpty(t, content)
}
