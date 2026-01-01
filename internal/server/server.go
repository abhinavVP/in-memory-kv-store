package server

import (
	"fmt"
	"io"
	"kvs/internal/command"
	"kvs/internal/execution"
	"kvs/internal/messaging"
	"log"
	"net"
)

type Server struct{
	PORT				int
	CommandsChannel 	chan command.Command
	Store				map[string][]byte
}

func (s *Server)InitServer(port int, chanSize int){
	s.PORT = port
	s.CommandsChannel = make(chan command.Command, chanSize)
	s.Store = make(map[string][]byte)
}

func (s *Server) Listen(){
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.PORT))
	if err != nil{
		log.Printf("[server] error trying to listen: %v", err)
		return
	}
	defer listener.Close()
	log.Printf("[server] now listening on PORT %v", s.PORT)

	go execution.Executor(s.CommandsChannel, s.Store)

	for {
		conn , err := listener.Accept()
		if err != nil{
			log.Printf("[server] error trying to accept connection from a client: %v", err)
			continue
		}

		go s.handleClient(conn)
	}
}

func (s *Server)handleClient(conn net.Conn){
	for{
		cmd, err := messaging.Read(conn)
		if err != nil{
			if err == io.EOF {
				log.Println("[client handler] client closed connection")
			} else {
				log.Printf("[client handler] error trying to read from the client: %v", err)
			}
			return
		}

		s.CommandsChannel <- cmd
		res := cmd.Response()
		packet := messaging.SerializeResponse(res)

		_, err = conn.Write(packet)
		if err != nil{
			log.Printf("[client handler] error trying to write message to the client: %v", err)
			continue
		}
	}
}


