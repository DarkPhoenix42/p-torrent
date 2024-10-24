package peer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const BitTorrentProtocolHeader = "\x13BitTorrent protocol"
const BitTorrentExtensions = "\x00\x00\x00\x00\x00\x00\x00\x00"

type Peer struct {
	Addr net.Addr
	Conn net.Conn

	ID         [20]byte
	Choked     bool
	Interested bool
	Bitfield   []byte
}

func NewPeer(addr net.Addr) Peer {
	return Peer{
		Addr:       addr,
		ID:         [20]byte{},
		Choked:     true,
		Interested: false,
		Bitfield:   make([]byte, 0),
	}
}

func (p *Peer) handShake(info_hash []byte, peer_id []byte) error {
	var err error

	p.Conn, err = net.DialTimeout("tcp", p.Addr.String(), 5*time.Second)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(BitTorrentProtocolHeader)
	buf.WriteString(BitTorrentExtensions)
	buf.Write(info_hash)
	buf.Write(peer_id)

	p.Conn.Write(buf.Bytes())

	response := make([]byte, 68)
	_, err = p.Conn.Read(response)
	if err != nil {
		return err
	}

	if string(response[:20]) != BitTorrentProtocolHeader {
		return fmt.Errorf("unexpected handshake response from peer %s", p.Addr.String())
	}

	if string(response[28:48]) != string(info_hash) {
		return fmt.Errorf("unexpected handshake response from peer %s", p.Addr.String())
	}

	copy(p.ID[:], response[48:68])
	return nil
}

func (p *Peer) recvLoop() error {
	msg_len := make([]byte, 4)
	for {
		_, err := io.ReadFull(p.Conn, msg_len)
		if err != nil {
			return err
		}

		msg_bytes := make([]byte, binary.BigEndian.Uint32(msg_len))
		_, err = io.ReadFull(p.Conn, msg_bytes)
		if err != nil {
			return err
		}

		msg := DeserialiseMessage(msg_bytes)

		switch msg.ID {

		case Choke:
			p.Choked = true

		case UnChoke:
			p.Choked = false

		case Interested:
			p.Interested = true

		case NotInterested:
			p.Interested = false

		case Have:
			// Don't care because we are only a client

		case Bitfield:
			p.Bitfield = msg.Payload
			fmt.Printf("Peer %s has bitfield %v\n", p.Addr.String(), msg.Payload)

		case Request:
			// Don't care because we are only a client

		case Piece:
			fmt.Printf("Peer %s sent piece %d\n", p.Addr.String(), binary.BigEndian.Uint32(msg.Payload))

		case Cancel:
			// Don't care because we are only a client

		default:
			fmt.Printf("unexpected message ID %d from peer %s\n", msg.ID, p.Addr.String())
		}

	}
}

func (p *Peer) Activate(info_hash []byte, peer_id []byte) error {
	err := p.handShake(info_hash, peer_id)
	if err != nil {
		return err
	}

	p.SendUnchoke()
	p.SendInterested()
	err = p.recvLoop()
	if err != nil {
		return err
	}
	return nil

}

func (p *Peer) SendUnchoke() {
	p.Conn.Write(SerialiseMessage(Message{
		ID:      UnChoke,
		Payload: []byte{},
	}))
}

func (p *Peer) SendInterested() {
	p.Conn.Write(SerialiseMessage(Message{
		ID:      Interested,
		Payload: []byte{},
	}))
}
