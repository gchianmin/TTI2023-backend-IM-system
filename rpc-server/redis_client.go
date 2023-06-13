package main

import (
	"context"
    "encoding/json"
    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
	cli *redis.Client
}

// create another struct for JSON serialization
type Message struct {
	Sender    string  `json:"sender"`
    Message   string  `json:"message"`
    Timestamp int64   `json:"timestamp"`
}

func (c *RedisClient) InitClient(ctx context.Context, address, password string) error {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, 
		DB:       0,        
    })

    if err := r.Ping(ctx).Err(); err != nil {
		return err
    }

    c.cli = r

    return nil
}

func (c *RedisClient) SaveMessage(ctx context.Context, roomID string, message *Message) error {
	
    chat, err := json.Marshal(message)
    if err != nil {
		return err
    }

    member := &redis.Z{
		Score: float64(message.Timestamp), 
		Member: chat, 
    }

    _, err = c.cli.ZAdd(ctx, roomID, *member).Result()
    if err != nil {
		return err
    }

    return nil
}

func (c *RedisClient) GetMessagesByRoomID(ctx context.Context, roomID string, start, end int64, reverse bool) ([]*Message, error) {
    var (
       rawMessages []string
       messages    []*Message
       err         error
    )

	// if the order is reversed, use ZRevRange instead of ZRANGE
    if reverse {
       rawMessages, err = c.cli.ZRevRange(ctx, roomID, start, end).Result()
       if err != nil {
          return nil, err
       }
    } else {
       rawMessages, err = c.cli.ZRange(ctx, roomID, start, end).Result()
       if err != nil {
          return nil, err
       }
    }

    for _, msg := range rawMessages {
       temp := &Message{}
       err := json.Unmarshal([]byte(msg), temp)
       if err != nil {
          return nil, err
       }
       messages = append(messages, temp)
    }

    return messages, nil
}