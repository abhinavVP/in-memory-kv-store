package messaging

import (
	"encoding/binary"
	"fmt"
	"io"
	"kvs/internal/command"
)

func Read(r io.Reader) (command.Command, error){
	header := make([]byte, 16)

	n, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}
	if n != 16 {
		return nil, fmt.Errorf("incomplete header read")
	}

	cmdType := header[0]
	flags := header[1]
	keyLen := binary.BigEndian.Uint16(header[2:4])
	valueLen := binary.BigEndian.Uint32(header[4:8])
	ttl := binary.BigEndian.Uint64(header[8:16])

	key := make([]byte, keyLen)
	_, err = io.ReadFull(r, key)
	if err != nil {
		return nil, err
	}
	switch cmdType {

	case 1:
		if valueLen == 0 {
			return nil, fmt.Errorf("SET command missing value")
		}

		value := make([]byte, valueLen)
		_, err = io.ReadFull(r, value)
		if err != nil {
			return nil, err
		}

		var ttlVal uint64
		if flags & 0x01 != 0 {
			ttlVal = ttl
		}

		return &command.Set{
			Key:    string(key),
			Value:  value,
			Result: make(chan command.Result),
			TTL:    ttlVal,
		}, nil

	case 2:
		return &command.Get{
			Key:    string(key),
			Result: make(chan command.Result),
		}, nil

	case 3:
		return &command.Delete{
			Key:    string(key),
			Result: make(chan command.Result),
		}, nil

	default:
		return nil, fmt.Errorf("received invalid command type: %d", cmdType)
	}
}	

func SerializeResponse(r command.Result) []byte {
	header := make([]byte, 5)

	header[0] = byte(r.Status)

	var valueLen uint32
	if r.Status == command.StatusOK && len(r.Message) > 0 {
		valueLen = uint32(len(r.Message))
	} else {
		valueLen = 0
	}

	binary.BigEndian.PutUint32(header[1:], valueLen)

	if valueLen == 0 {
		return header
	}

	packet := make([]byte, 5+valueLen)
	copy(packet[:5], header)
	copy(packet[5:], r.Message)

	return packet
}

