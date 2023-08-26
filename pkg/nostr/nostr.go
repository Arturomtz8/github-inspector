package nostr

import (
	"bytes"
	"context"
	"log"
	"os"
	"sync"
	"text/template"

	"github.com/nbd-wtf/go-nostr"

	"github.com/Arturomtz8/github-inspector/pkg/github"
)

// PusblishRepos function get the repos info,
// parse them and publish them to Nostr relays concurrently.
func PusblishRepos() error {
	wg := sync.WaitGroup{}

	repos, err := github.SearchGithubTrending("Go", github.RepoURL)
	if err != nil {
		return err
	}

	reposContent, err := getReposContent(repos)
	if err != nil {
		return err
	}

	wg.Add(len(reposContent))
	for _, repo := range reposContent {
		go func(repo string) {
			defer wg.Done()
			publishRepo(repo)
		}(repo)
	}

	wg.Wait()

	return nil
}

// getReposContent function parse repos into a template.
func getReposContent(repos *github.TrendingSearchResult) ([]string, error) {
	tmplFile := "repo.tmpl"
	reposContent := make([]string, 0)

	for _, repo := range repos.Items {
		buf := &bytes.Buffer{}

		tmpl, err := template.New(tmplFile).ParseFiles(tmplFile)
		if err != nil {
			return nil, err
		}
		err = tmpl.Execute(buf, repo)
		if err != nil {
			return nil, err
		}

		reposContent = append(reposContent, buf.String())
	}

	return reposContent, nil
}

// publishRepo publish content to a Nostr Relay, get the private key from
// an environment variable.
func publishRepo(content string) error {
	sk := os.Getenv("NOSTR_HEX_SK")
	pub, _ := nostr.GetPublicKey(sk)

	ev := nostr.Event{
		PubKey:    pub,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      nil,
		Content:   content,
	}

	// calling Sign sets the event ID field and the event Sig field
	ev.Sign(sk)

	// publish the event to two relays
	ctx := context.Background()
	for _, url := range []string{"wss://nostr.danvergara.com", "wss://relay.danvergara.com"} {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			return err
		}
		_, err = relay.Publish(ctx, ev)
		if err != nil {
			return err
		}

		log.Printf("published to %s\n", url)
	}

	return nil
}