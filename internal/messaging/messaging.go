package messaging

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"kvs/internal/command"
	"kvs/internal/execution"
)

func Read(buf *bufio.Reader, header []byte) (execution.Request, error){
	_, err := io.ReadFull(buf, header)
	if err != nil {
		return execution.Request{}, err
	}

	t := header[0]
	// flags := header[1]
	len_key := binary.BigEndian.Uint16(header[2:4])
	len_val := binary.BigEndian.Uint32(header[4:8])
	ttl := binary.BigEndian.Uint64(header[8:16])

	payload := make([]byte, uint32(len_key)+len_val)
	_, err = io.ReadFull(buf, payload)
	if err != nil {
		return execution.Request{}, err
	}

	key := payload[:len_key]

	switch t {
	case 1:
		if len_val==0 {
			return execution.Request{}, fmt.Errorf("[client reader] didnt get value arg for SET command")
		}

		req := execution.Request{
			Cmd: command.Command{
				Ctype: command.SET,
				Key: key,
				Value: payload[len_key:len_val],
				Ttl: ttl,
			},
		}

		return req, nil
	case 2:
		req :=  execution.Request{
			Cmd: command.Command{
				Ctype: command.GET,
				Key: key,
			},
		}

		return req, nil
	case 3:
		req := execution.Request{
			Cmd: command.Command{
				Ctype: command.DELETE,
				Key: key,
			},
		}

		return req, nil
	default:
		return execution.Request{}, fmt.Errorf("[client reader] invalid client request")
	}
}

func Write(buf *bufio.Writer, header []byte, r command.Result) error {
	header[0] = byte(r.Status)
	len_val := len(r.Value)
	binary.BigEndian.PutUint32(header[1:], uint32(len_val))
	_, err := buf.Write(header)
	if err != nil{
		return err
	}
	if (len_val>0){
		buf.Write(r.Value)
	}

	return nil

}