package torrent

import (
	"crypto/sha1"
	"os"

	"github.com/DarkPhoenix42/p-torrent/pkg/bencode"
)

type Torrent struct {
	InfoHash [20]byte
	Info     Info
	Announce string
}

type Info struct {
	Name        string
	PieceLength int
	Pieces      [][20]byte
	Length      int
	Files       []File
}

type File struct {
	Length int
	Path   []string
}

func newFile(m map[string]any) File {
	f := File{}
	for key, value := range m {
		switch key {
		case "length":
			f.Length = value.(int)
		case "path":
			for _, path := range value.([]any) {
				f.Path = append(f.Path, path.(string))
			}
		}
	}
	return f
}

func marshallableFile(f File) map[string]any {
	m := map[string]any{
		"length": f.Length,
		"path":   []any{},
	}
	for _, path := range f.Path {
		m["path"] = append(m["path"].([]any), path)
	}
	return m
}

func newInfo(m map[string]any) Info {
	info := Info{}
	for key, value := range m {
		switch key {
		case "name":
			info.Name = value.(string)
		case "piece length":
			info.PieceLength = value.(int)
		case "pieces":
			piecesStr := value.(string)
			info.Pieces = make([][20]byte, len(piecesStr)/20)
			for i := 0; i < len(piecesStr); i += 20 {
				copy(info.Pieces[i/20][:], piecesStr[i:i+20])
			}
		case "length":
			info.Length = value.(int)
		case "files":
			for _, file := range value.([]any) {
				info.Files = append(info.Files, newFile(file.(map[string]any)))
			}
		}
	}
	return info
}

func marshallableInfo(info Info) map[string]any {
	m := map[string]any{
		"name":         info.Name,
		"piece length": info.PieceLength,
		"pieces":       []byte{},
	}

	for _, piece := range info.Pieces {
		m["pieces"] = append(m["pieces"].([]byte), piece[:]...)
	}
	if len(info.Files) > 0 {
		m["files"] = []any{}
	} else {
		m["length"] = info.Length
	}
	for _, file := range info.Files {
		m["files"] = append(m["files"].([]any), marshallableFile(file))
	}
	return m
}

func NewTorrent(filename string) (*Torrent, error) {
	file_data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewTorrentFromBencode(file_data)
}

func NewTorrentFromBencode(bencoded []byte) (*Torrent, error) {
	unmarshalled_data, err := bencode.UnMarshal(bencoded)
	if err != nil {
		return nil, err
	}

	t := &Torrent{}
	for key, value := range unmarshalled_data.(map[string]any) {
		switch key {
		case "info":
			t.Info = newInfo(value.(map[string]any))
		case "announce":
			t.Announce = value.(string)
		}
	}

	err = t.updateInfoHash()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (torrent *Torrent) updateInfoHash() error {

	info := marshallableInfo(torrent.Info)
	info_bencoded, err := bencode.Marshal(info)
	if err != nil {
		return err
	}

	torrent.InfoHash = sha1.Sum(info_bencoded)
	return nil
}

func (torrent *Torrent) GetLength() int {
	if len(torrent.Info.Files) > 0 {
		length := 0
		for _, file := range torrent.Info.Files {
			length += file.Length
		}
		return length
	}

	return torrent.Info.Length
}
