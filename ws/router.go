package main

import (
	"encoding/json"
	"log"
)

type HandlerFunc func(client *Client, payload json.RawMessage)

type Router struct {
	handlers map[MessageType]HandlerFunc
}

func NewRouter() *Router {
	return &Router{handlers: make(map[MessageType]HandlerFunc)}
}

func (r *Router) On(msgType MessageType, fn HandlerFunc) {
	r.handlers[msgType] = fn
}

func (r *Router) Dispatch(client *Client, msg WireMessage) {
	handler, ok := r.handlers[msg.Type]
	if !ok {
		log.Printf("no handler for message type: %s", msg.Type)
		return
	}
	handler(client, msg.Payload)
}
