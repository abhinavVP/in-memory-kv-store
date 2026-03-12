package server

import (
	"container/heap"
	"fmt"
	"io"
	"kvs/internal/command"
	"kvs/internal/execution"
	"kvs/internal/messaging"
	"kvs/internal/ttlheap"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct{
	PORT				int
	CommandsChannel 	chan command.Command
	Store				map[string][]byte
	TTLMap				map[string]uint64
	TTLheap				*ttlheap.ExpiryHeap
	mu					sync.Mutex
}

func (s *Server)InitServer(port int, chanSize int){
	h := &ttlheap.ExpiryHeap{}
	heap.Init(h)

	s.PORT = port
	s.CommandsChannel = make(chan command.Command, chanSize)
	s.Store = make(map[string][]byte)
	s.TTLMap = make(map[string]uint64)
	s.TTLheap = h
}

func (s *Server) Listen(){
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.PORT))
	if err != nil{
		log.Printf("[server] error trying to listen: %v", err)
		return
	}
	defer listener.Close()
	log.Printf("[server] now listening on PORT %v", s.PORT)

	go execution.Executor(s.CommandsChannel, s.Store, s.TTLMap, *s.TTLheap)

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

func (s *Server) TTLMonitor() {
	for {
		s.mu.Lock()

		if s.TTLheap.Len() == 0 {
			s.mu.Unlock()
			time.Sleep(time.Second)
			continue
		}

		item := (*s.TTLheap)[0]
		now := uint64(time.Now().Unix())

		if item.ExpiresAt > now {
			sleepFor := time.Duration(item.ExpiresAt-now) * time.Second
			s.mu.Unlock()
			time.Sleep(sleepFor)
			continue
		}

		expiredItem := heap.Pop(s.TTLheap).(*ttlheap.ExpiryItem)

		currentExpiry, ok := s.TTLMap[expiredItem.Key]
		if !ok || currentExpiry != expiredItem.ExpiresAt {
			s.mu.Unlock()
			continue
		}

		delete(s.TTLMap, expiredItem.Key)
		delete(s.Store, expiredItem.Key)

		s.mu.Unlock()
	}
}



