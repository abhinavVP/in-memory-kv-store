package execution

import (
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

func (e *Executor) Run(){

	for r := range e.Reqch {
		switch (r.Cmd.Ctype){
		case command.SET:
			r.Resch <- command.ExecSet(e.Store, r.Cmd)
		case command.GET:
			r.Resch <- command.ExecGet(e.Store, r.Cmd)
		case command.DELETE:
			r.Resch <- command.ExecDelete(e.Store, r.Cmd)
		}
			
	}
}