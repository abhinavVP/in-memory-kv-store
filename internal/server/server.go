package server

import (
	"bufio"
	"context"
	"hash/fnv"
	"kvs/internal/command"
	"kvs/internal/execution"
	"kvs/internal/messaging"
	"kvs/internal/persistance"
	"log"
	"net"
	"sync"
)

type Server struct{
	PORT	string
	IP		string
	shards	[]*execution.Executor
	N		int
	AOFpath string
}

type Client struct{
	Conn 	net.Conn
	Resch 	chan command.Result
}

func (s *Server)Init(port, ip string, n int, aofpath string){
	s.PORT = port
	s.IP = ip
	s.N = n
	s.AOFpath = aofpath
}

func (s *Server)Listen(ctx context.Context){
	lis, err := net.Listen("tcp", net.JoinHostPort(s.IP, s.PORT))
	if err != nil{
		log.Printf("[server] error trying to listen: %v", err)
		return
	}

	defer lis.Close()

	aof := persistance.AOF{}
	go aof.Work(s.AOFpath, ctx)

	s.shards = make([]*execution.Executor, s.N)
	for i := range s.shards{
		s.shards[i] = &execution.Executor{
			Reqch: make(chan execution.Request, 1024),
			Store: make(map[string][]byte),
			Mu: sync.Mutex{},
		}

		go s.shards[i].Run(ctx, aof.Commands)
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

		go s.clientReader(c, ctx)
		go s.clientWriter(c, ctx)
	}
}

func (s *Server)clientReader(c Client, ctx context.Context){
	buf := bufio.NewReader(c.Conn)
	header := make([]byte, 16)

	go func(){
		<- ctx.Done()
		c.Conn.Close()
	} ()

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

func (s *Server)clientWriter(c Client, ctx context.Context){
	buf := bufio.NewWriter(c.Conn)
	header := make([]byte, 5)
	for{
		select{
		case <- ctx.Done():
			return
		case r, ok := <- c.Resch:
			if !ok {
				return
			}
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
}

func (s *Server) getShard(key []byte) *execution.Executor {
    h := fnv.New32a()
    h.Write(key)
    return s.shards[h.Sum32() % uint32(len(s.shards))]
}