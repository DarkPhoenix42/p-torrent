package peer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/DarkPhoenix42/p-torrent/pkg/piece"
)

const (
	BitTorrentProtocolHeader = "\x13BitTorrent protocol"
	BitTorrentExtensions     = "\x00\x00\x00\x00\x00\x00\x00\x00"
	ReadTimeout              = 15
	DialTimeout              = 10

	BlockReqLength   = int(1 << 14)
	MaxPendingBlocks = 25
)

type Peer struct {
	Addr net.Addr
	Conn net.Conn

	ID         [20]byte
	Choked     bool
	Interested bool
	BitField   BitField

	Work            chan *piece.Piece
	Results         chan *piece.Piece
	PieceInProgress *piece.Piece

	Responsive    bool
	PendingBlocks int
	Downloaded    int
}

func NewPeer(addr net.Addr, work, results chan *piece.Piece, numPieces int) *Peer {
	return &Peer{
		Addr: addr,
		Conn: nil,

		ID:       [20]byte{},
		Choked:   true,
		BitField: make([]byte, (numPieces+7)/8),

		Work:            work,
		Results:         results,
		PieceInProgress: nil,

		Responsive:    true,
		PendingBlocks: 0,
		Downloaded:    0,
	}
}

func (p *Peer) HandShake(info_hash []byte, peer_id []byte) error {
	var err error

	p.Conn, err = net.DialTimeout("tcp", p.Addr.String(), DialTimeout*time.Second)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(BitTorrentProtocolHeader)
	buf.WriteString(BitTorrentExtensions)
	buf.Write(info_hash)
	buf.Write(peer_id)

	_, err = p.Conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	response := make([]byte, 68)
	p.Conn.SetReadDeadline(time.Now().Add(ReadTimeout * time.Second))
	defer p.Conn.SetReadDeadline(time.Time{})

	_, err = io.ReadFull(p.Conn, response)

	if err != nil {
		return err
	}

	if string(response[:20]) != BitTorrentProtocolHeader {
		return fmt.Errorf("[%s] Unexpected handshake response.", p.Addr.String())
	}

	if string(response[28:48]) != string(info_hash) {
		return fmt.Errorf("[%s] Unexpected handshake response.", p.Addr.String())
	}

	copy(p.ID[:], response[48:68])
	return nil
}

func (p *Peer) Activate(info_hash []byte, peer_id []byte) error {
	err := p.HandShake(info_hash, peer_id)
	if err != nil {
		return err
	}

	err = p.SendUnchoke()
	if err != nil {
		return err
	}
	err = p.SendInterested()
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) HandleIncoming() error {
	msg_len := make([]byte, 4)
	for {
		p.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		_, err := io.ReadFull(p.Conn, msg_len)
		p.Conn.SetReadDeadline(time.Time{})

		if err != io.EOF && err != nil {
			fmt.Printf("[%s] Error reading message len: %s\n", p.Addr, err)
			if p.PieceInProgress != nil {
				p.Work <- p.PieceInProgress
				p.PieceInProgress = nil
			}

			p.Responsive = false
			return err
		}

		msg_bytes := make([]byte, binary.BigEndian.Uint32(msg_len))

		p.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		_, err = io.ReadFull(p.Conn, msg_bytes)
		p.Conn.SetReadDeadline(time.Time{})

		if err != io.EOF && err != nil {
			fmt.Printf("[%s] Error reading message bytes: %s\n", p.Addr, err)
			if p.PieceInProgress != nil {
				p.Work <- p.PieceInProgress
				p.PieceInProgress = nil
			}
			p.Responsive = false
			return err
		}

		msg := DeserialiseMessage(msg_bytes)
		switch msg.ID {
		case MsgChoke:
			p.Choked = true

		case MsgUnChoke:
			p.Choked = false

		case MsgHave:
			piece_index := int(binary.BigEndian.Uint32(msg.Payload))
			p.BitField.SetPiece(piece_index)

		case MsgPiece:
			err := ParsePieceMessage(msg, p.PieceInProgress)

			if err != nil {
				fmt.Printf("Error parsing piece message from %s : %s\n", p.Addr, err)
				p.Work <- p.PieceInProgress
				p.PieceInProgress = nil
				continue
			}

			p.Downloaded += BlockReqLength
			p.PendingBlocks--

			if p.Downloaded >= p.PieceInProgress.Length && p.PendingBlocks == 0 {
				if p.PieceInProgress.Validate() {
					p.Results <- p.PieceInProgress
				} else {
					fmt.Printf("[%s] Invalid piece #%d\n", p.Addr, p.PieceInProgress.Index)
					p.Work <- p.PieceInProgress
				}

				p.Downloaded = 0
				p.PieceInProgress = nil
			}

		case MsgBitfield:
			copy(p.BitField, msg.Payload)
		}

	}
}

func (p *Peer) DownloadPiece() error {
	p.PendingBlocks = 0
	requested := 0

	fmt.Printf("[%s] Requested piece #%d\n", p.Addr, p.PieceInProgress.Index)

	for p.PendingBlocks < MaxPendingBlocks && requested < p.PieceInProgress.Length && !p.Choked {
		blockSize := BlockReqLength

		if p.PieceInProgress.Length-requested < blockSize {
			blockSize = p.PieceInProgress.Length - requested
		}
		err := p.SendPieceRequest(p.PieceInProgress.Index, requested, blockSize)

		if err != nil {
			return err
		}

		p.PendingBlocks++
		requested += blockSize
	}

	return nil
}

func (p *Peer) StartDownload() {
	go p.HandleIncoming()

	for {
		if p.Choked || p.PieceInProgress != nil {
			continue
		}

		if !p.Responsive {
			return
		}

		p.PieceInProgress = <-p.Work
		if p.BitField.HasPiece(p.PieceInProgress.Index) {
			err := p.DownloadPiece()

			if err != nil {
				fmt.Printf("[%s] Error downloading piece #%d : %s\n", p.Addr, p.PieceInProgress.Index, err)
				p.Work <- p.PieceInProgress
				p.PieceInProgress = nil
			}

		} else {
			p.Work <- p.PieceInProgress
			p.PieceInProgress = nil
		}

	}
}
