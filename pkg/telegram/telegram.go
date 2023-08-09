package telegram

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"text/template"

	"github.com/Arturomtz8/github-inspector/pkg/github"
)

const (
	startCommand string = "/start"
	RepoURL             = "https://api.github.com/search/repositories"

	telegramApiBaseUrl     string = "https://api.telegram.org/bot"
	telegramApiSendMessage string = "/sendMessage"
	telegramTokenEnv       string = "GITHUB_BOT_TOKEN"
)

var lenStartCommand int = len(startCommand)

type Chat struct {
	Id int `json:"id"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

func HandleTelegramWebhook(_ http.ResponseWriter, r *http.Request) {
	var update, err = parseTelegramRequest(r)
	if err != nil {
		log.Printf("error parsing update, %s", err.Error())
		return
	}

	var sanitizedString = sanitize(update.Message.Text)

	result, err := github.SearchGithubTrending(sanitizedString)
	if err != nil {
		log.Fatal("Error with searchgithubtrending", err)
	}

	const templ = `{{.TotalCount}} repositories:
	{{range .Items}}----------------------------------------
	Name:          {{.Full_name}}
	Url:           {{.Html_url}}
	Description:   {{.Description}}
	Created at:    {{.Created_at }}
	Update _at:    {{.Updated_at}} 
	Pushed at:     {{.Pushed_at}}
	Size(KB):      {{.Size}}
	Language:      {{.Language}}
	Stargazers:    {{.Stargazers_count}}
	Archived:      {{.Archived}}
	Open Issues:   {{.Open_issues_count}}
	Topics:        {{.Topics}}
	{{end}}`

	var report = template.Must(template.New("trendinglist").Parse(templ))
	buf := &bytes.Buffer{}
	if err := report.Execute(buf, result); err != nil {
		panic(err)
	}

	s := buf.String()
	// testString := "hello"
	var telegramResponseBody, errTelegram = sendTextToTelegramChat(update.Message.Chat.Id, s)
	if errTelegram != nil {
		log.Printf("got error %s from telegram, response body is %s", errTelegram.Error(), telegramResponseBody)

	} else {
		log.Printf("successfully distributed to chat id %d", update.Message.Chat.Id)
	}

}

func parseTelegramRequest(r *http.Request) (*Update, error) {
	var update Update

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("could not decode incoming update %s", err.Error())
		return nil, err
	}
	return &update, nil
}

func sanitize(s string) string {
	if len(s) >= lenStartCommand {
		if s[:lenStartCommand] == startCommand {
			s = s[lenStartCommand:]
		}
	}
	return s

}

func sendTextToTelegramChat(chatId int, text string) (string, error) {
	log.Printf("Sending %s to chat_id: %d", text, chatId)

	var telegramApi string = "https://api.telegram.org/bot" + os.Getenv("GITHUB_BOT_TOKEN") + "/sendMessage"

	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {text},
		})
	if err != nil {
		log.Printf("error when posting text to the chat: %s", err.Error())
		return "", err
	}
	// defer response.Body.Close()
	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("error parsing telegram answer %s", errRead.Error())
		return "", err
	}

	bodyString := string(bodyBytes)
	log.Printf("body of telegram response: %s", bodyString)
	response.Body.Close()
	return bodyString, nil

}
