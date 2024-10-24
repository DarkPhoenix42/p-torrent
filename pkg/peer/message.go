package peer

import (
	"bytes"
	"encoding/binary"
)

type MsgType int

const (
	Choke MsgType = iota
	UnChoke
	Interested
	NotInterested
	Have
	Bitfield
	Request
	Piece
	Cancel
)

type Message struct {
	ID      MsgType
	Payload []byte
}

func SerialiseMessage(msg Message) []byte {
	var buf bytes.Buffer
	msg_len := uint32(len(msg.Payload) + 1)
	binary.Write(&buf, binary.BigEndian, msg_len)
	buf.WriteByte(byte(msg.ID))
	buf.Write(msg.Payload)
	return buf.Bytes()
}

func DeserialiseMessage(data []byte) Message {
	msg := Message{}
	msg.ID = MsgType(data[0])
	msg.Payload = data[1:]
	return msg
}

func HasPiece(bitfield []byte, piece_index int) bool {
	byte_index := piece_index / 8
	offset := piece_index % 8
	return bitfield[byte_index]>>(7-offset)&1 != 0
}

