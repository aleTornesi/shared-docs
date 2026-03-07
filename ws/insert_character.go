package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type Insert struct {
	Text string `json:"text"`
}

var redisClient = redis.NewClient(&redis.Options{
	Addr: "redis:6379",
})

func GetUserCursorPosition(c *Client) (string, error) {
	key := fmt.Sprintf("%d:%d", c.room, c.userID)
	return redisClient.Get(context.Background(), key).Result()
}

func SetUserCursorPosition(c *Client, position string) error {
	key := fmt.Sprintf("%d:%d", c.room, c.userID)
	return redisClient.Set(context.Background(), key, position, 0).Err()
}

func insertText(c *Client, insertData *Insert) {
	val, err := GetUserCursorPosition(c)
	if err != nil {
		log.Printf("GetUserCursorPosition: %v", err)
		return
	}
	c.send <- []byte(val)
}
