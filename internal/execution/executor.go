package execution

import (
	"kvs/internal/command"
	"log"
)

func Executor(commands chan command.Command, items map[string][]byte){
	log.Print("[executor] started")
	for cmd := range commands{
		cmd.Execute(items)	
	}
	log.Printf("[executor] channel closed, stopping executor...")
}