package nostr

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"

	"github.com/Arturomtz8/github-inspector/pkg/github"
)

// PusblishRepos function get the repos info,
// parse them and publish them to Nostr relays.
func PusblishRepos(ctx context.Context, sk, redisURI string) error {
	// Makes a request every 8 secs/8000 miliseconds,
	// since most relays have strict rate limits.
	limiter := rate.NewLimiter(rate.Every(8000*time.Millisecond), 1)

	repos, err := github.GetTrendingRepos(github.TimeToday, "Go")
	if err != nil {
		return err
	}

	filteredRepos, err := filterReposBasedKeys(ctx, redisURI, repos.Items)
	if err != nil {
		return err
	}

	for _, repo := range filteredRepos {
		if err := limiter.Wait(ctx); err != nil {
			continue
		}

		tmplRepo, err := tmplRepocontent(repo)
		if err != nil {
			// No need to break loop, just continue to the next one.
			log.Printf("error occurred parsing repo into template: %v", err)
			continue
		}

		log.Printf("repo: %s", tmplRepo)
		if err := publishRepo(tmplRepo, sk); err != nil {
			// No need to break loop, just continue to the next one.
			log.Printf("error occurred publishing repo: %v", err)
			continue
		}
	}

	return nil
}

func filterReposBasedKeys(ctx context.Context, redisURI string, repos []*github.RepoTrending) ([]*github.RepoTrending, error) {
	var filteredRepos []*github.RepoTrending

	// redisURI := "rediss://default:AVNS_FMPwK8_PVi8JbXhNE_U@redis-nostr-awesome-projects-daniel-bb11.aivencloud.com:21705"

	addr, err := redis.ParseURL(redisURI)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(addr)

	for _, repo := range repos {
		repoKey, err := rdb.Get(ctx, repo.FullName).Result()
		if err == redis.Nil {
			log.Printf("%s is not seen", repo.FullName)

			// the repos is not seen, so the repo can be published.
			filteredRepos = append(filteredRepos, repo)

			// Store the key for 36 hrs.
			// 36 * 60 * 60 = 129600.
			err = rdb.Set(ctx, repo.FullName, repo.HtmlURL, time.Second*129600).Err()
			if err != nil {
				log.Printf("error occurred storing repo's full name: %v", err)
				continue
			}
		} else if err != nil {
			// No need to break loop, just continue to the next one.
			log.Printf("error occurred publishing getting repo from redis: %v", err)
			continue
		} else {
			log.Printf("%s is seen and can safely be skipped", repoKey)
			continue
		}
	}

	return filteredRepos, nil
}

// tmplRepocontent function parse repos into a template.
func tmplRepocontent(repo *github.RepoTrending) (string, error) {
	tmplFile := "repo.tmpl"

	buf := &bytes.Buffer{}

	tmpl, err := template.New(tmplFile).ParseFiles(tmplFile)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(buf, repo)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// publishRepo publish content to a Nostr Relay, get the private key from
// an environment variable.
func publishRepo(content, sk string) error {
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
