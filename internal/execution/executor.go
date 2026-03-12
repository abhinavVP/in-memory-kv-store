package execution

import (
	"kvs/internal/command"
	"kvs/internal/ttlheap"
	"log"
	"sync"
)

func Executor(commands chan command.Command, items map[string][]byte, TTLmap map[string]uint64, TTLheap *ttlheap.ExpiryHeap, mu *sync.Mutex){
	log.Print("[executor] started")
	for cmd := range commands{
		cmd.Execute(items, TTLmap, TTLheap, mu)	
	}
	log.Printf("[executor] channel closed, stopping executor...")
}