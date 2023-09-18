package nostr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Arturomtz8/github-inspector/pkg/github"
)

func TestParseRepoContent(t *testing.T) {

	repo := &github.RepoTrending{
		FullName:        "foo/bar",
		HtmlURL:         "https://github.com/foo/bar",
		Description:     "a good project",
		Language:        "Go",
		StargazersCount: 1000,
		Owner: github.Owner{
			Login: "foo",
		},
	}

	parsedRepo, err := tmplRepocontent(repo)
	t.Log(parsedRepo)
	require.NoError(t, err)
	require.NotEmpty(t, parsedRepo)
}
