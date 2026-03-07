package main

import "sync"

type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	mu         sync.RWMutex
}

type Message struct {
	room    string
	payload []byte
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.room]; !ok {
				h.rooms[client.room] = make(map[*Client]bool)
			}
			h.rooms[client.room][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.room]; ok {
				delete(room, client)
				close(client.send)
				if len(room) == 0 {
					delete(h.rooms, client.room)
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.rooms[msg.room] {
				select {
				case client.send <- msg.payload:
				default:
					close(client.send)
					delete(h.rooms[msg.room], client)
				}
			}
			h.mu.RUnlock()
		}
	}
}
