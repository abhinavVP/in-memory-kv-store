package persistance

import (
	"bufio"
	"context"
	"encoding/binary"
	"kvs/internal/command"
	"log"
	"os"
	"time"
)

type AOF struct{
	Commands	chan command.Command
	Buf			*bufio.Writer
}

func appendToBuffer(c command.Command, buf *bufio.Writer){
	buf.WriteByte(byte(c.Ctype))
    buf.WriteByte(0) //for flags idk what to do yet since i dont really have a any type of flags created yet other than ttl
    binary.Write(buf, binary.BigEndian, uint16(len(c.Key)))
    binary.Write(buf, binary.BigEndian, uint32(len(c.Value)))
    binary.Write(buf, binary.BigEndian, c.Ttl)
    buf.Write(c.Key)
    buf.Write(c.Value)
}

func (a *AOF) Work(path string, ctx context.Context){
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil{
		log.Printf("[aof] error opening aof file: %v", err)
		return
	}
	
	defer f.Close()
	a.Buf = bufio.NewWriter(f)

	timer := time.NewTicker(time.Second)
	for {
		select {
		case cmd := <- a.Commands:
			appendToBuffer(cmd, a.Buf)
		case <- timer.C:
			a.Buf.Flush()
			f.Sync()
		case <- ctx.Done():
			a.Buf.Flush()
			f.Sync()
			return
		}
	}

}
