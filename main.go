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
	url := fmt.Sprintf("https://api.github.com/repos/tokopedia/ios-tokopedia/pulls?access_token=%v&head=tokopedia:%v", githubToken, branchName)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

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

func getSlackUser(email string) (*User, error) {
	url := fmt.Sprintf("https://slack.com/api/users.lookupByEmail?token=%v&email=%v", slackToken, email)
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

	if !slackResponse.Ok || slackResponse.User == nil {
		return nil, fmt.Errorf("user %v does not exist", email)
	}

	return slackResponse.User, nil
}

func generateAttachments(user User) []interface{} {

	attachments := []interface{}{
		map[string]interface{}{
			"fallback":    "Required plain-text summary of the attachment.",
			"color":       "#2eb886",
			"pretext":     "Optional text that appears above the attachment block",
			"author_name": "Bobby Tables",
			"author_link": "http://flickr.com/bobby/",
			"author_icon": "http://flickr.com/icons/bobby.jpg",
			"title":       "Slack API Documentation",
			"title_link":  "https://api.slack.com/",
			"text":        "Optional text that appears within the attachment",
			"image_url":   "http://my-website.com/path/to/image.jpg",
			"thumb_url":   "http://example.com/path/to/thumb.png",
			"footer":      "Slack API",
			"footer_icon": "https://platform.slack-edge.com/img/default_application_icon.png",
			"ts":          123456789,
		},
	}

	return attachments
}

func sendMessage(user User) error {

	type message struct {
		Channel     string        `json:"channel"`
		Text        string        `json:"text"`
		AsUser      bool          `json:"as_user"`
		Attachments []interface{} `json:"attachments"`
	}

	m := message{
		Channel:     user.ID,
		AsUser:      true,
		Attachments: generateAttachments(user),
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", slackToken))

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

	data := make(map[string]interface{})
	err = json.Unmarshal(b, &data)
	if err != nil {
		log.Println(err)
		return
	}

	notes, ok := data["notes"].(string)
	if !ok {
		log.Println("error convert notes to string")
		return
	}

	branch, err := getBranch(notes)
	if err != nil {
		log.Println(err)
		return
	}

	prInfo, err := getPRInfo(branch)
	if err != nil {
		log.Println(err)
		return
	}

	reviewerEmails := getReviewerEmails(prInfo.Body)
	for _, email := range reviewerEmails {

		user, err := getSlackUser(email)
		if err != nil {
			log.Println(err)
			continue
		}

		err = sendMessage(*user)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("message sent to: %v\n", user.RealName)
	}
}

func main() {

	githubToken = flag.String("githubToken", "", "Github Token")
	if githubToken == nil {
		log.Fatalln("Please provide githubToken")
	}

	slackToken = flag.String("slackToken", "", "Slack Token")
	if slackToken == nil {
		log.Fatalln("Please provide slackToken")
	}

	http.HandleFunc("/webhook-handler", webhookHandler)

	log.Fatal(http.ListenAndServe(":8888", nil))
}
