package main

type GithubUser struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

type PRInfo struct {
	Body       string     `json:"body"`
	GithubUser GithubUser `json:"user"`
}

type SlackUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

type SlackResponse struct {
	Ok        bool       `json:"ok"`
	SlackUser *SlackUser `json:"user"`
	Error     string     `json:"error"`
}

type AppVersion struct {
	ShortVersion string `json:"shortversion"`
	Notes        string `json:"notes"`
}

type WebhookRequest struct {
	Text       string     `json:"text"`
	AppVersion AppVersion `json:"app_version"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Fallback   string  `json:"fallback"`
	Color      string  `json:"color"`
	AuthorName string  `json:"author_name"`
	AuthorLink string  `json:"author_link"`
	AuthorIcon string  `json:"author_icon"`
	Pretext    string  `json:"pretext"`
	Text       string  `json:"text"`
	Fields     []Field `json:"fields"`
	ImageURL   string  `json:"image_url"`
	ThumbURL   string  `json:"thumb_url"`
	Footer     string  `json:"footer"`
	FooterIcon string  `json:"footer_icon"`
}

type SlackMessage struct {
	Channel     string       `json:"channel"`
	Text        string       `json:"text"`
	AsUser      bool         `json:"as_user"`
	Attachments []Attachment `json:"attachments"`
}
