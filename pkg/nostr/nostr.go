package nostr

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"golang.org/x/time/rate"

	"github.com/Arturomtz8/github-inspector/pkg/github"
)

// defaulRepoLen defaults to 10 repos to be publish, it can be less if
// the length of the response is smaller.
const defaulRepoLen = 10

// PusblishRepos function get the repos info,
// parse them and publish them to Nostr relays.
func PusblishRepos(ctx context.Context) error {
	var repoLen int

	// Makes a request every 6 secs/6000 miliseconds,
	// since most relays have strict rate limits.
	limiter := rate.NewLimiter(rate.Every(6000*time.Millisecond), 1)

	repos, err := github.GetTrendingRepos(github.TimeToday, "Go")
	if err != nil {
		return err
	}

	reposContent, err := getReposContent(repos)
	if err != nil {
		return err
	}

	// if the length of the response is smaller,
	// then add that value to repoLen,
	// otherwise go with default.
	if len(reposContent) <= defaulRepoLen {
		repoLen = len(reposContent)
	} else {
		repoLen = defaulRepoLen
	}

	for i := 0; i < repoLen; i++ {
		if err := limiter.Wait(ctx); err != nil {
			continue
		}
		repo := reposContent[i]
		log.Printf("repo: %s", repo)
		if err := publishRepo(repo); err != nil {
			// No need to break loop, just continue to the next one.
			log.Printf("error occurred publishing event %v", err)
			continue
		}
	}

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
	for _, url := range []string{
		"wss://nostr.danvergara.com",
		"wss://relay.damus.io/",
		"wss://relay.nostr.band",
	} {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			return err
		}
		status, err := relay.Publish(ctx, ev)
		if err != nil {
			return fmt.Errorf("error publishing event %v with status %d", err, status)
		}

		log.Printf("published to %s\n", url)
	}

	return nil
}
