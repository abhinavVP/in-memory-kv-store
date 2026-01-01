package command

type CommandType uint8

const (
	_ CommandType = iota
	SET
	GET
	DELETE
)

type Command interface{
	Execute(map[string][]byte)
	Response()	Result
}

type Result struct{
	Type	CommandType
	Success	bool
	Message []byte
}


type Set struct{
	Key		string
	Value	[]byte
	Result	chan Result
}

type Get struct{
	Key		string
	Result	chan Result
}

type Delete struct{
	Key		string
	Result	chan Result
}

func (s *Set) Execute(items map[string][]byte) {
	items[s.Key] = s.Value
	s.Result <- Result{
		Type: 1,
		Success: true,
		Message: nil,
	}
}

func (g *Get) Execute(items map[string][]byte){
	value, ok := items[g.Key]
	if ok {
		g.Result <- Result{
			Type: 2,
			Success: true,
			Message: value,
		}
	}else{
		g.Result <- Result{
			Type: 2,
			Success: false,
			Message: nil,
		}
	}
}

func (d *Delete) Execute(items map[string][]byte) {
	delete(items, d.Key)

	d.Result <- Result{
		Type: 3,
		Success: true,
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




