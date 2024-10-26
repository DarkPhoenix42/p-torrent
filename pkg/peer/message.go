package peer

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/DarkPhoenix42/p-torrent/pkg/piece"
)

type MsgType int

const (
	MsgChoke MsgType = iota
	MsgUnChoke
	MsgInterested
	MsgNotInterested
	MsgHave
	MsgBitfield
	MsgRequest
	MsgPiece
	MsgCancel
	MsgKeepAlive
)

type Message struct {
	ID      MsgType
	Payload []byte
}

func (msg *Message) Serialise() []byte {
	var buf bytes.Buffer
	msg_len := uint32(len(msg.Payload) + 1)
	binary.Write(&buf, binary.BigEndian, msg_len)
	buf.WriteByte(byte(msg.ID))
	buf.Write(msg.Payload)
	return buf.Bytes()
}

func DeserialiseMessage(data []byte) Message {
	if len(data) == 0 {
		return Message{ID: MsgKeepAlive}
	}
	msg := Message{}
	msg.ID = MsgType(data[0])
	msg.Payload = data[1:]
	return msg
}

type BitField []byte

func (bitfield BitField) HasPiece(piece_index int) bool {
	byte_index := piece_index / 8
	offset := piece_index % 8
	return bitfield[byte_index]>>uint(7-offset)&1 != 0
}

func (bitfield BitField) SetPiece(piece_index int) {
	byte_index := piece_index / 8
	offset := piece_index % 8
	bitfield[byte_index] |= 1 << uint(7-offset)
}

func (p *Peer) SendUnchoke() error {
	msg := Message{
		ID: MsgUnChoke,
	}

	_, err := p.Conn.Write(msg.Serialise())
	if err != nil {
		return err
	}
	return nil
}

func (p *Peer) SendInterested() error {
	msg := Message{
		ID: MsgInterested,
	}
	_, err := p.Conn.Write(msg.Serialise())
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) SendPieceRequest(index, begin, length int) error {
	msg := Message{
		ID:      MsgRequest,
		Payload: make([]byte, 12),
	}

	binary.BigEndian.PutUint32(msg.Payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(msg.Payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(msg.Payload[8:12], uint32(length))

	_, err := p.Conn.Write(msg.Serialise())
	if err != nil {
		return err
	}

	return nil
}

func ParsePieceMessage(msg Message, piece *piece.Piece) error {
	index := binary.BigEndian.Uint32(msg.Payload[0:4])
	begin := binary.BigEndian.Uint32(msg.Payload[4:8])

	if int(index) != piece.Index {
		return fmt.Errorf("expected piece index %d, got %d", piece.Index, index)
	}

	copy(piece.Data[begin:], msg.Payload[8:])

	return nil
}
