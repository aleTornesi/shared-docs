package main

import (
	"encoding/json"
	"testing"
)

func TestRouterOnAndDispatch(t *testing.T) {
	router := NewRouter()
	called := false
	var gotPayload json.RawMessage

	router.On(MessageTypeInsertText, func(client *Client, payload json.RawMessage) {
		called = true
		gotPayload = payload
	})

	payload := json.RawMessage(`{"text":"a"}`)
	msg := WireMessage{Type: MessageTypeInsertText, Payload: payload}
	router.Dispatch(&Client{}, msg)

	if !called {
		t.Fatal("handler not called")
	}
	if string(gotPayload) != string(payload) {
		t.Fatalf("payload %s, want %s", gotPayload, payload)
	}
}

func TestRouterDispatchUnknown(t *testing.T) {
	router := NewRouter()
	// Should not panic
	router.Dispatch(&Client{}, WireMessage{Type: "unknown_type", Payload: nil})
}
