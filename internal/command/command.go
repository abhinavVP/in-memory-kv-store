package command

import (
	"kvs/internal/ttlheap"
	"time"
)

type CommandType uint8
type StatusCode uint8

const (
	_ CommandType = iota
	SET
	GET
	DELETE
)

const (
	StatusOK StatusCode = iota
	StatusKeyNotFound
	StatusKeyExistsButNoValue
	StatusInvalidCommand
	StatusInternalError
)

type Command interface{
	Execute(map[string][]byte, map[string]uint64, ttlheap.ExpiryHeap)
	Response()	Result
}

type Result struct{
	Status	StatusCode
	Message	[]byte
}


type Set struct{
	Key		string
	Value	[]byte
	Result	chan Result
	TTL		uint64
}

type Get struct{
	Key		string
	Result	chan Result
}

type Delete struct{
	Key		string
	Result	chan Result
}

func (s *Set) Execute(items map[string][]byte, TTLmap map[string]uint64, TTLheap ttlheap.ExpiryHeap) {
	items[s.Key] = s.Value
	if s.TTL > 0{
		expiry := time.Now().Add(time.Duration(s.TTL) * time.Second).Unix()
		TTLmap[s.Key] = uint64(expiry)
	}
	s.Result <- Result{
		Status: 0,
		Message: nil,
	}
}

func (g *Get) Execute(items map[string][]byte, TTLmap map[string]uint64, TTLheap ttlheap.ExpiryHeap){
	value, ok := items[g.Key]
	if ok {
		g.Result <- Result{
			Status: 0,
			Message: value,
		}
	}else{
		g.Result <- Result{
			Status: 1,
			Message: nil,
		}
	}
}

func (d *Delete) Execute(items map[string][]byte, TTLmap map[string]uint64, TTLheap ttlheap.ExpiryHeap) {
	delete(items, d.Key)

	d.Result <- Result{
		Status: 0,
		Message: nil,
	}
}

func (s *Set) Response() Result{
	return <- s.Result
}

func (g *Get) Response() Result{
	return <- g.Result
}

func (d *Delete) Response() Result{
	return <- d.Result
}