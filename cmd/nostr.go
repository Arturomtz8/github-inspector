/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/Arturomtz8/github-inspector/pkg/nostr"
)

// nostrCmd represents the nostr command
var nostrCmd = &cobra.Command{
	Use:   "nostr",
	Short: "Publish Go Repos to Nostr Relays",
	Long:  `Publish Go Repos data to Nostr Relays`,
	Run: func(_ *cobra.Command, _ []string) {
		sk := os.Getenv("NOSTR_HEX_SK")
		redisURL := os.Getenv("REDIS_HOST")
		redisPassword := os.Getenv("REDIS_PASSWORD")
		ctx := context.Background()
		if err := nostr.PusblishRepos(ctx, sk, redisURL, redisPassword); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(nostrCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nostrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nostrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
