package piece

import "crypto/sha1"

type Piece struct {
	Index  int
	Length int
	Hash   [20]byte
	Data   []byte
}

func NewPiece(index int, length int, hash [20]byte) *Piece {
	return &Piece{
		Index:  index,
		Length: length,
		Hash:   hash,
		Data:   make([]byte, length),
	}
}

func (p *Piece) Validate() bool {
	return sha1.Sum(p.Data) == p.Hash
}
