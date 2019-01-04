package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	githubToken *string
	slackToken  *string
)

func getBranch(notes string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(notes))
	if err != nil {
		return "", err
	}

	return doc.Find("p:first-child").Text(), nil
}

func getPRInfo(branchName string) (*PRInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/tokopedia/ios-tokopedia/pulls?access_token=%v&head=tokopedia:%v", *githubToken, branchName)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	log.Println(url)

	var listOfPrInfo []*PRInfo
	err = json.NewDecoder(res.Body).Decode(&listOfPrInfo)
	if err != nil {
		return nil, err
	}

	if listOfPrInfo == nil || len(listOfPrInfo) == 0 {
		return nil, fmt.Errorf("branch %v does not exist", branchName)
	}

	return listOfPrInfo[0], nil
}

func getReviewerEmails(prInfoBody string) []string {
	reg := regexp.MustCompile("([a-zA-Z0-9._-]+@tokopedia.com)")
	return reg.FindAllString(prInfoBody, len(prInfoBody))
}

func getSlackUser(email string) (*SlackUser, error) {
	url := fmt.Sprintf("https://slack.com/api/users.lookupByEmail?token=%v&email=%v", *slackToken, email)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var slackResponse SlackResponse
	err = json.NewDecoder(res.Body).Decode(&slackResponse)
	if err != nil {
		return nil, err
	}

	if !slackResponse.Ok || slackResponse.SlackUser == nil {
		return nil, fmt.Errorf("user %v does not exist", email)
	}

	return slackResponse.SlackUser, nil
}

func sendMessage(slackMessage SlackMessage) error {

	b, err := json.Marshal(slackMessage)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", *slackToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var slackResponse SlackResponse
	err = json.NewDecoder(res.Body).Decode(&slackResponse)
	if err != nil {
		return err
	}

	if !slackResponse.Ok {
		return errors.New(slackResponse.Error)
	}

	return nil
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var webhookRequest WebhookRequest
	err = json.Unmarshal(b, &webhookRequest)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%+v\n", webhookRequest)

	branch, err := getBranch(webhookRequest.AppVersion.Notes)
	if err != nil {
		log.Println(err)
		return
	}

	prInfo, err := getPRInfo(branch)
	if err != nil {
		log.Println(err)
		return
	}

	slackMessage := SlackMessage{
		AsUser: true,
		Attachments: []Attachment{
			Attachment{
				Fallback:   "Please kindly review / check my latest app.",
				Pretext:    webhookRequest.Text,
				Color:      "#2eb886",
				AuthorName: prInfo.GithubUser.Login,
				AuthorLink: prInfo.GithubUser.HTMLURL,
				AuthorIcon: prInfo.GithubUser.AvatarURL,
				Text:       "Please kindly review / check my latest app.",
				Fields: []Field{
					{
						Title: "Version",
						Value: webhookRequest.AppVersion.ShortVersion,
						Short: true,
					},
					{
						Title: "Branch",
						Value: branch,
						Short: true,
					},
				},
				ImageURL:   "https://ecs7.tokopedia.net/blog-tokopedia-com/uploads/2015/08/tokopedia.png",
				ThumbURL:   "https://ecs.tokopedia.com/img/footer/toped.png",
				Footer:     "PokeTheReviewer",
				FooterIcon: "https://ecs.tokopedia.com/img/footer/toped.png",
			},
		},
	}

	reviewerEmails := getReviewerEmails(prInfo.Body)
	for _, email := range reviewerEmails {
		go func(email string, slackMessage SlackMessage) {
			slackUser, err := getSlackUser(email)
			if err != nil {
				log.Println(err)
				return
			}

			slackMessage.Channel = slackUser.ID
			err = sendMessage(slackMessage)
			if err != nil {
				log.Println(err)
				return
			}

			log.Printf("message sent to: %v\n", slackUser.RealName)
		}(email, slackMessage)
	}
}

func main() {
	githubToken = flag.String("githubToken", "", "Github Token")
	slackToken = flag.String("slackToken", "", "Slack Token")

	flag.Parse()
	if githubToken == nil || *githubToken == "" {
		log.Fatalln("Please provide githubToken")
	}
	log.Printf("github token = %v\n", *githubToken)

	if slackToken == nil || *slackToken == "" {
		log.Fatalln("Please provide slackToken")
	}
	log.Printf("slack token = %v\n", *slackToken)
	flag.Parse()

	http.HandleFunc("/webhook-handler", webhookHandler)

	port := ":8888"
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("PokeTheReviewer is running on port %v\n", port)
}
