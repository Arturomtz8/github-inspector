package github_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Arturomtz8/github-inspector/pkg/github"
)

func TestSearchGithubTrending(t *testing.T) {
	testDataFile, _ := os.Open("testdata/response.json")
	testDataFileContent, _ := io.ReadAll(testDataFile)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(testDataFileContent)
	}))

	results, err := github.SearchGithubTrending("go", s.URL)
	require.NoError(t, err)
	require.Len(t, results.Items, 5)
}
