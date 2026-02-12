package database

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedClient(host, port, password string, useTLS bool)(*redis.Client, error) {
	opts := &redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB: 0,
	}
	if useTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil , err
	}

	return client , nil
}

