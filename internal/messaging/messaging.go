package messaging

import (
	"encoding/binary"
	"fmt"
	"io"
	"kvs/internal/command"
)

func Read(r io.Reader) (command.Command, error){
	header := make([]byte, 6)

	n, err := io.ReadFull(r, header)

	if err != nil{
		return nil, err
	}
	if n == 0{
		return nil, fmt.Errorf("read 0 bytes from client, unable to read")
	}

	t := header[0] >> 6 // first 2 bit give the command type, 1 = SET, 2 = GET, 3 = DELETE
	len_key := binary.BigEndian.Uint16(header[:2]) & 0x3FFF
	key := make([]byte, len_key)
	_, err = io.ReadFull(r, key)
	if err != nil{
		return nil, nil
	}
	switch t{
	case 0:
		return nil, fmt.Errorf("recieved invalid type from client")
	case 1:
		len_value := binary.BigEndian.Uint32(header[2:])
		value := make([]byte, len_value)
		_, err = io.ReadFull(r, value)
		if err != nil{
			return nil, err
		}
		return &command.Set{
			Key: string(key),
			Value: value,
			Result: make(chan command.Result),
		}, nil
			
	case 2:
		return &command.Get{
			Key: string(key),
			Result: make(chan command.Result),
		}, nil
	case 3:
		return &command.Delete{
			Key: string(key),
			Result: make(chan command.Result),
		}, nil
	default:
		return nil, fmt.Errorf("recieved invalid type from client")			
	}
}	

func SerializeResponse(r command.Result) []byte{
	header := make([]byte, 5)
	t := byte(r.Type)
	header[0] = t
	
	switch t{
	case 1,3:
		return header
	case 2:
		if r.Success { //if key exists
			len_value := uint32(len(r.Message))
			binary.BigEndian.PutUint32(header[1:], len_value)
			packet := make([]byte,5 + len_value)
			copy(packet[0:5], header)

			if len_value > 0{ // sometimes can be empty value
				value := r.Message
				copy(packet[5:], value)
			}
			return packet	
		}else{
			len_value := uint32(0xFFFFFFFF)
			binary.BigEndian.PutUint32(header[1:], len_value)
			
			return header
		}
		
	default:
		return nil
	}
}