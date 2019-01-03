package main

type PRInfo struct {
	Body string `json:"body"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

type SlackResponse struct {
	Ok    bool   `json:"ok"`
	User  *User  `json:"user"`
	Error string `json:"error"`
}
