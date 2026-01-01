package main

import "kvs/internal/server"

func main(){
	s := &server.Server{}
	s.InitServer(12345, 7)
	s.Listen()
}
