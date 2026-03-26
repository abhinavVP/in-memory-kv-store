package server

import (
	"bufio"
	"hash/fnv"
	"kvs/internal/command"
	"kvs/internal/execution"
	"kvs/internal/messaging"
	"log"
	"net"
	"sync"
)

type Server struct{
	PORT	string
	IP		string
	shards	[]*execution.Executor
	N		int
}

type Client struct{
	Conn 	net.Conn
	Resch 	chan command.Result
}

func (s *Server)Init(port, ip string, n int){
	s.PORT = port
	s.IP = ip
	s.N = n
}

func (s *Server)Listen(){
	lis, err := net.Listen("tcp", net.JoinHostPort(s.IP, s.PORT))
	if err != nil{
		log.Printf("[server] error trying to listen: %v", err)
		return
	}

	defer lis.Close()

	s.shards = make([]*execution.Executor, s.N)
	for i := range s.shards{
		s.shards[i] = &execution.Executor{
			Reqch: make(chan execution.Request, 1024),
			Store: make(map[string][]byte),
			Mu: sync.Mutex{},
		}

		go s.shards[i].Run()
	}

	for {
		conn , err := lis.Accept()
		if err != nil{
			log.Printf("[server] error trying to accept connection from a client: %v", err)
			continue
		}

		c := Client{
			Conn: conn,
			Resch: make(chan command.Result, 7),
		}

		go s.clientReader(c)
		go s.clientWriter(c)
	}
}

func (s *Server)clientReader(c Client){
	buf := bufio.NewReader(c.Conn)
	header := make([]byte, 16)
	for{	
		r, err := messaging.Read(buf, header)
		if err != nil{
			log.Print(err.Error())
			c.Conn.Close()
			close(c.Resch)
			return
		}
		r.Resch = c.Resch
		shard := s.getShard(r.Cmd.Key)
		shard.Reqch <- r
	}
}
func (s *Server)clientWriter(c Client){
	buf := bufio.NewWriter(c.Conn)
	header := make([]byte, 5)
	for{
		r := <- c.Resch
		err := messaging.Write(buf, header, r)
		if err != nil{
			log.Printf("[client writer] %v", err.Error())
			continue
		}

		if len(c.Resch) == 0{
			err = buf.Flush()
			if err != nil{
				log.Printf("[client writer] %v", err.Error())
			}
		}
	}
}

func (s *Server) getShard(key []byte) *execution.Executor {
    h := fnv.New32a()
    h.Write(key)
    return s.shards[h.Sum32() % uint32(len(s.shards))]
}