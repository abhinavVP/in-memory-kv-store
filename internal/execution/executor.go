package execution

import (
	"context"
	"kvs/internal/command"
	"sync"
)

type Request struct{
	Cmd		command.Command
	Resch	chan command.Result
}

type Executor struct{
	Reqch	chan Request
	Store	map[string][]byte
	Mu		sync.Mutex
}

func (e *Executor) Run(ctx context.Context, aof chan command.Command){

	for{
		select{
		case r := <- e.Reqch:
			switch (r.Cmd.Ctype){
			case command.SET:
				aof <- r.Cmd
				r.Resch <- command.ExecSet(e.Store, r.Cmd)
			case command.GET:
				r.Resch <- command.ExecGet(e.Store, r.Cmd)
			case command.DELETE:
				aof <- r.Cmd
				r.Resch <- command.ExecDelete(e.Store, r.Cmd)
			}
		case <- ctx.Done():
			close(e.Reqch)
			return
		}	
	}
}