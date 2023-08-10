package cmd

import (
	"fmt"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/spf13/cobra"

	"github.com/Arturomtz8/github-inspector/pkg/telegram"
)

// telegramCmd represents the telegram command.
var telegramCmd = &cobra.Command{
	Use:   "telegram",
	Short: "telegram commands that runs a web hook which responds with trending projects on GitHub",
	Long: `Telegram commands that runs a web hook which responds with trending projects on GitHub.
  Originally intended to run on Google Cloud Functions.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("telegram bot called")

		functions.HTTP("HandleTelegramWebhook", telegram.HandleTelegramWebhook)
	},
}

func init() {
	rootCmd.AddCommand(telegramCmd)
}
