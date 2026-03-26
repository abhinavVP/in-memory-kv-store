package main

import "kvs/internal/server"

func main(){
	s := &server.Server{}
	s.Init("12345", "127.0.0.1", 4)
	s.Listen()
}
