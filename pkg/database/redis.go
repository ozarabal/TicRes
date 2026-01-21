package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedClient(host, port, password string)(*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB: 0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil , err
	}

	return client , nil
}

