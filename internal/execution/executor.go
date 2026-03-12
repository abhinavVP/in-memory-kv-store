package execution

import (
	"kvs/internal/command"
	"kvs/internal/ttlheap"
	"log"
)

func Executor(commands chan command.Command, items map[string][]byte, TTLmap map[string]uint64, TTLheap ttlheap.ExpiryHeap){
	log.Print("[executor] started")
	for cmd := range commands{
		cmd.Execute(items, TTLmap, TTLheap)	
	}
	log.Printf("[executor] channel closed, stopping executor...")
}