package nostr

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
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

func TestFilteredRepos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping short mode")
	}

	ctx := context.Background()

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// simulates an scenario where repos can be duplicated.
	repos := []*github.RepoTrending{
		{
			FullName: "danvergara/dblab",
			HtmlURL:  "https://github.com/danvergara/dblab",
		},
		{
			FullName: "danvergara/dblab",
			HtmlURL:  "https://github.com/danvergara/dblab",
		},
		{
			FullName: "tofu/tofu",
			HtmlURL:  "https://github.com/tofu/tofu",
		},
		{
			FullName: "tofu/tofu",
			HtmlURL:  "https://github.com/tofu/tofu",
		},
		{
			FullName: "argoproj/argo-workflows",
			HtmlURL:  "https://github.com/argoproj/argo-workflows",
		},
		{
			FullName: "go-resty/resty",
			HtmlURL:  "https://github.com/go-resty/resty",
		},
	}

	filteredRepos, err := filterReposBasedKeys(ctx, redisAddr, redisPassword, repos)

	require.NoError(t, err)
	// the lenght of the repos slice should be 4 because there are only 4 unique repos.
	require.Len(t, filteredRepos, 4)

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0, // use default DB
	})
	// add duplicated repos to redis.
	rdb.Set(ctx, "argoproj/argo-workflows", "https://github.com/argoproj/argo-workflows", time.Second*2).Err()
	rdb.Set(ctx, "go-resty/resty", "https://github.com/go-resty/resty", time.Second*2).Err()

	filteredRepos, err = filterReposBasedKeys(ctx, redisAddr, redisPassword, repos)
	require.NoError(t, err)
	// the repo should be empty since there's no new repo.
	require.Len(t, filteredRepos, 0)

	// let the keys expire.
	time.Sleep(time.Second * 2)

	filteredRepos, err = filterReposBasedKeys(ctx, redisAddr, redisPassword, repos)
	require.NoError(t, err)
	// should be 2 repos since dblab and tofu have longer expire times.
	require.Len(t, filteredRepos, 2)
}
