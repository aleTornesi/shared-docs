package main

import "encoding/json"

type MessageType string

const (
	MessageTypeInsertText MessageType = "insert_character"
	MessageTypeDeleteText MessageType = "delete_character"
	MessageTypeMoveCursor MessageType = "move_cursor"
)

// WireMessage is the JSON envelope received/sent over WebSocket.
type WireMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
