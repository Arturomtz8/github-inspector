package github

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchGithubTrending(t *testing.T) {
	testDataFile, _ := os.Open("testdata/response.json")
	testDataFileContent, _ := io.ReadAll(testDataFile)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(testDataFileContent)
	}))

	results, err := SearchGithubTrending("go", s.URL)
	require.NoError(t, err)
	require.Len(t, results.Items, 5)
	require.Equal(t, results.Items[0].FullName, "kriskowal/q")
	require.Equal(t, results.Items[0].Owner.Login, "kriskowal")
	require.Equal(t, results.Items[0].Description, "A promise library for JavaScript")
	require.Equal(t, results.Items[0].StargazersCount, 14955)
}

func TestEncodeQueryComponents(t *testing.T) {
	type input struct {
		repo   string
		lang   string
		author string
	}
	type expected struct {
		output string
		hasErr bool
	}
	var tests = []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name:  "empty name",
			input: input{},
			expected: expected{
				hasErr: true,
			},
		},
		{
			name: "just passing repository name",
			input: input{
				repo: "dblab",
			},
			expected: expected{
				output: "q=dblab+in%3Aname",
			},
		},
		{
			name: "providing a language",
			input: input{
				repo: "dblab",
				lang: "go",
			},
			expected: expected{
				output: "q=dblab+in%3Aname+language%3Ago",
			},
		},
		{
			name: "providing author",
			input: input{
				repo:   "dblab",
				author: "danvergara",
			},
			expected: expected{
				output: "q=dblab+in%3Aname+user%3Adanvergara",
			},
		},
		{
			name: "providing all",
			input: input{
				repo:   "dblab",
				author: "danvergara",
				lang:   "go",
			},
			expected: expected{
				output: "q=dblab+in%3Aname+user%3Adanvergara+language%3Ago",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := encodeQueryComponents(tc.input.repo, tc.input.lang, tc.input.author)
			if tc.expected.hasErr {
				require.Error(t, err)
				return
			}

			require.Equal(t, tc.expected.output, got)
		})
	}
}

func TestGetRepository(t *testing.T) {
	type input struct {
		repo   string
		lang   string
		author string
	}
	type expected struct {
		output *RepoTrending
		hasErr bool
	}
	var tests = []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name: "providing all parameters",
			input: input{
				repo:   "dblab",
				author: "danvergara",
				lang:   "go",
			},
			expected: expected{
				output: &RepoTrending{
					FullName:    "danvergara/dblab",
					Description: "The database client every command line junkie deserves.",
					Owner: Owner{
						Login: "danvergara",
					},
					Language: "Go",
				},
			},
		},
		{
			name: "providing WRONG parameters",
			input: input{
				repo:   "dblab",
				author: "wrong-author",
				lang:   "go",
			},
			expected: expected{
				hasErr: true,
			},
		},
	}
	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			// Create a server that returns a static JSON response.
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.expected.hasErr {
					w.WriteHeader(http.StatusUnprocessableEntity)
					w.Write([]byte(`{
            "message": "Validation Failed",
            "errors": [
              {
                "message": "The listed users and repositories cannot be searched either because the resources do not exist or you do not have permission to view them.",
                "resource": "Search",
                "field": "q",
                "code": "invalid"
              }
            ],
  "         documentation_url": "https://docs.github.com/v3/search/"
          }`))
					return
				} else {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
            "total_count": 1,
            "incomplete_results": false,
            "items": [
              {
                "id": 352879222,
                "node_id": "MDEwOlJlcG9zaXRvcnkzNTI4NzkyMjI=",
                "name": "dblab",
                "full_name": "danvergara/dblab",
                "private": false,
                "owner": {
                  "login": "danvergara",
                  "id": 12239167,
                  "node_id": "MDQ6VXNlcjEyMjM5MTY3",
                  "avatar_url": "https://avatars.githubusercontent.com/u/12239167?v=4",
                  "gravatar_id": "",
                  "url": "https://api.github.com/users/danvergara",
                  "html_url": "https://github.com/danvergara",
                  "followers_url": "https://api.github.com/users/danvergara/followers",
                  "following_url": "https://api.github.com/users/danvergara/following{/other_user}",
                  "gists_url": "https://api.github.com/users/danvergara/gists{/gist_id}",
                  "starred_url": "https://api.github.com/users/danvergara/starred{/owner}{/repo}",
                  "subscriptions_url": "https://api.github.com/users/danvergara/subscriptions",
                  "organizations_url": "https://api.github.com/users/danvergara/orgs",
                  "repos_url": "https://api.github.com/users/danvergara/repos",
                  "events_url": "https://api.github.com/users/danvergara/events{/privacy}",
                  "received_events_url": "https://api.github.com/users/danvergara/received_events",
                "type": "User",
                "site_admin": false
                },
                "html_url": "https://github.com/danvergara/dblab",
                "description": "The database client every command line junkie deserves.",
                "fork": false,
                "url": "https://api.github.com/repos/danvergara/dblab",
                "forks_url": "https://api.github.com/repos/danvergara/dblab/forks",
                "keys_url": "https://api.github.com/repos/danvergara/dblab/keys{/key_id}",
                "collaborators_url": "https://api.github.com/repos/danvergara/dblab/collaborators{/collaborator}",
                "teams_url": "https://api.github.com/repos/danvergara/dblab/teams",
                "hooks_url": "https://api.github.com/repos/danvergara/dblab/hooks",
                "issue_events_url": "https://api.github.com/repos/danvergara/dblab/issues/events{/number}",
                "created_at": "2021-03-30T05:20:36Z",
                "updated_at": "2023-09-13T08:51:23Z",
                "pushed_at": "2023-08-15T23:38:38Z",
                "size": 14815,
                "stargazers_count": 702,
                "watchers_count": 702,
                "language": "Go"
              }
            ]
          }`))
					return
				}
			}))

			// Be sure to clean up the server or you might run out of file descriptors!
			t.Cleanup(func() {
				s.Close()
			})

			got, err := GetRepository(s.URL, tc.input.repo, tc.input.lang, tc.input.author)
			if tc.expected.hasErr {
				require.Error(t, err)
				return
			}

			require.Equal(t, tc.expected.output.FullName, got.FullName)
			require.Equal(t, tc.expected.output.Description, got.Description)
			require.Equal(t, tc.expected.output.Owner.Login, got.Owner.Login)
			require.Equal(t, tc.expected.output.Language, got.Language)
		})
	}
}
