package command

import (

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

type Command struct{
    Ctype   CommandType
    Key     []byte
    Value   []byte
    Ttl     uint64
}

type Result struct{
    Status StatusCode
    Value  []byte
}

func ExecSet(store map[string][]byte, cmd Command) Result{
    store[string(cmd.Key)] = cmd.Value

    return Result{
        Status: StatusOK,
    }
}

func ExecGet(store map[string][]byte, cmd Command) Result{
    val, ok := store[string(cmd.Key)]
    if !ok {
        return Result{
            Status: StatusKeyNotFound,
        }
    } 
    if len(val) == 0{
        return Result{
            Status: StatusKeyExistsButNoValue,
        }
    } else {
        return Result{
            Status: StatusOK,
            Value: val,
        }
    }
}

func ExecDelete(store map[string][]byte, cmd Command) Result{
    delete(store, string(cmd.Key))

    return Result{
        Status: StatusOK,
    }
}