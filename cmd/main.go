package main

import (
	"context"
	"kvs/internal/server"
)

func main(){
	s := &server.Server{}
	s.Init("12345", "127.0.0.1", 4, "")
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	s.Listen(ctx, )
}
