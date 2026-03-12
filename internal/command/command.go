package command

import (
	"container/heap"
	"kvs/internal/ttlheap"
	"sync"
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
	Execute(map[string][]byte, map[string]uint64, *ttlheap.ExpiryHeap, *sync.Mutex)
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

func (s *Set) Execute(items map[string][]byte, TTLmap map[string]uint64, TTLheap *ttlheap.ExpiryHeap, mu *sync.Mutex) {
    mu.Lock()
    defer mu.Unlock() // Ensures the lock is released when the function exits

    items[s.Key] = s.Value
    if s.TTL > 0 {
        expiry := time.Now().Add(time.Duration(s.TTL) * time.Second).Unix()
        TTLmap[s.Key] = uint64(expiry)

        item := &ttlheap.ExpiryItem{
            Key:       s.Key,
            ExpiresAt: uint64(expiry),
        }
        
        heap.Push(TTLheap, item)
    } else {
        // If overwriting a key without a new TTL, remove any old TTL record
        delete(TTLmap, s.Key)
    }

    s.Result <- Result{
        Status: StatusOK,
        Message: nil,
    }
}

func (g *Get) Execute(items map[string][]byte, TTLmap map[string]uint64, TTLheap *ttlheap.ExpiryHeap, mu *sync.Mutex) {
    mu.Lock()
    defer mu.Unlock()

    value, ok := items[g.Key]
    if ok {
        if expiry, hasTTL := TTLmap[g.Key]; hasTTL {
            now := uint64(time.Now().Unix())
            
            if now >= expiry {
                // Passive expiration: delete from both maps
                delete(items, g.Key)
                delete(TTLmap, g.Key)
                
                g.Result <- Result{
                    Status:  StatusKeyNotFound,
                    Message: nil,
                }
                return
            }
        }
        status := StatusOK
        if len(value) == 0 {
            status = StatusKeyExistsButNoValue
        }
        g.Result <- Result{
            Status: status,
            Message: value,
        }
    } else {
        g.Result <- Result{
            Status: StatusKeyNotFound,
            Message: nil,
        }
    }
}

func (d *Delete) Execute(items map[string][]byte, TTLmap map[string]uint64, TTLheap *ttlheap.ExpiryHeap, mu *sync.Mutex) {
    mu.Lock()
    defer mu.Unlock()

    delete(items, d.Key)
    delete(TTLmap, d.Key) // Ensure we also clean up the TTL map

    d.Result <- Result{
        Status: StatusOK,
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