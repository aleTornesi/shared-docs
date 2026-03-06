package main

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type Page struct {
	Number  int16  `json:"number"`
	Content string `json:"content"`
}

type Document struct {
	ID     uint   `json:"id"`
	Title  string `json:"title"`
	Owner  User   `json:"owner"`
	Length uint   `json:"length,omitempty"`
	Pages  []Page `json:"pages,omitempty"`
}
