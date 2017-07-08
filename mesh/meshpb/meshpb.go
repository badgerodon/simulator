package meshpb

import (
	"encoding/binary"
	"io"

	"github.com/golang/protobuf/proto"
)

//go:generate protoc meshpb.proto --go_out=plugins=grpc:.

// Read reads a length-prefixed protobuf message from the reader
func Read(r io.Reader, msg proto.Message) error {
	sz := make([]byte, 8)
	_, err := io.ReadFull(r, sz)
	if err != nil {
		return err
	}

	bs := make([]byte, binary.BigEndian.Uint64(sz))
	_, err = io.ReadFull(r, bs)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(bs, msg)
	if err != nil {
		return err
	}
	return nil
}

// Write writes a length-prefixed protobuf message to the writer
func Write(w io.Writer, msg proto.Message) error {
	bs, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	sz := make([]byte, 8)
	binary.BigEndian.PutUint64(sz, uint64(len(bs)))
	_, err = w.Write(sz)
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}
