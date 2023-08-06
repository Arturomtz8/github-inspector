package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

const startCommand string = "/start"
const RepoURL = "https://api.github.com/search/repositories"

var lenStartCommand int = len(startCommand)

const telegramApiBaseUrl string = "https://api.telegram.org/bot"
const telegramApiSendMessage string = "/sendMessage"
const telegramTokenEnv string = "GITHUB_BOT_TOKEN"

var telegramApi string = telegramApiBaseUrl + os.Getenv(telegramTokenEnv) + telegramApiSendMessage

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

func init() {
	// Register an HTTP function with the Functions Framework
	functions.HTTP("HandleTelegramWebhook", HandleTelegramWebhook)
}

func HandleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	var update, err = parseTelegramRequest(r)
	if err != nil {
		log.Printf("error parsing update, %s", err.Error())
		return
	}

	var sanitizedString = sanitize(update.Message.Text)
	fmt.Println("String was sanitized")
	fmt.Println(sanitizedString)
	result, err := SearchGithubTrending(sanitizedString)

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
	log.Println(os.Getenv("GITHUB_BOT_TOKEN"))
	log.Println(chatId)
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

type TrendingSearchResult struct {
	TotalCount int
	Items      []*RepoTrending
}

type RepoTrending struct {
	Full_name         string
	Html_url          string //`json:"html_url"`
	Description       string
	Created_at        time.Time //`json:"created_at"`
	Updated_at        time.Time //`json:"updated_at"`
	Pushed_at         time.Time //`json:"pushed_at"`
	Size              int
	Language          string
	Stargazers_count  int
	Forks_count       int
	Archived          bool
	Open_issues_count int
	Topics            []string
}

func SearchGithubTrending(term string) (*TrendingSearchResult, error) {
	// in case receiving more values, consider changing to slice term string[]
	// q := url.QueryEscape(strings.Join(terms, " "))
	term = url.QueryEscape(term)
	// https://api.github.com/search/issues?q=stress+test+label:bug+language:python+state:closed&per_page=100
	fmt.Println("querying github api")
	resp, err := http.Get(RepoURL + "?q=stars:<=500+archived:false+language:" + term + "&per_page=5&sort=stars&order=desc")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("search query failed: %s", resp.Status)

	}

	var result TrendingSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body.Close()
	result.TotalCount = len(result.Items)
	return &result, nil
}
