package main

import "github.com/gorilla/websocket"

type Client struct {
	ws     *websocket.Conn
	send   chan []byte
	room   int64
	userID int64
	hub    *Hub
}
