package kernel

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Method types
const (
	MethodSocket = iota + 1
	MethodDial
	MethodListen
	MethodClose
)

func decodeSocketRequest(data []byte) {
}

func decodeSocketResult(data []byte) (handle int32) {
	decode(data, &handle)
	return
}

func decodeDialRequest(data []byte) (handle int32, port int32) {
	decode(data, &handle, &port)
	return
}

func decodeDialResult(data []byte) (err error) {
	decode(data, &err)
	return
}

func decodeListenRequest(data []byte) (handle int32, port int32) {
	decode(data, &handle, &port)
	return
}

func decodeListResult(data []byte) (err error) {
	decode(data, &err)
	return
}

func decodeCloseRequest(data []byte) (handle int32) {
	decode(data, &handle)
	return
}

func decodeCloseResult(data []byte) (err error) {
	decode(data, &err)
	return
}

func encodeSocketRequest() []byte {
	return encode(MethodSocket)
}

func encodeSocketResult(handle int32, err error) []byte {
	return encode(MethodSocket, handle, err)
}

func decode(data []byte, args ...interface{}) []byte {
	for _, arg := range args {
		switch t := arg.(type) {
		case *byte:
			*t = data[0]
			data = data[1:]
		case *int32:
			*t = int32(binary.LittleEndian.Uint32(data))
			data = data[4:]
		case *error:
			var str string
			data = decode(data, &str)
			if str == "" {
				*t = nil
			} else {
				*t = errors.New(str)
			}
		case *string:
			var sz int32
			data = decode(data, &sz)
			*t = string(data[:sz])
		default:
			panic(fmt.Sprintf("Unknown Type: %T", t))
		}
	}
	return data
}

func encode(args ...interface{}) []byte {
	var data []byte
	for _, arg := range args {
		switch t := arg.(type) {
		case byte:
			data = append(data, t)
		case int32:
			var tmp [4]byte
			binary.LittleEndian.PutUint32(tmp[:], uint32(t))
			data = append(data, tmp[:]...)
		case error:
			if t == nil {
				data = append(data, encode("")...)
			} else {
				data = append(data, encode(t.Error())...)
			}
		case string:
			data = append(data, encode(int32(len(t)))...)
			data = append(data, t...)
		}
	}
	return data
}
