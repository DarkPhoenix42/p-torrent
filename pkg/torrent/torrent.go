package torrent

import (
	"fmt"
	"os"

	"github.com/DarkPhoenix42/p-torrent/pkg/bencode"
)

type Torrent struct {
	InfoHash     [20]byte
	Info         Info
	Announce     string
	AnnounceList []string
	Comment      string
	CreatedBy    string
	CreationDate uint64
	Encoding     string
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

func NewFile(m map[string]any) File {
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

func NewInfo(m map[string]any) Info {
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
			fmt.Printf("%T\n", value)
			for _, file := range value.([]any) {
				info.Files = append(info.Files, NewFile(file.(map[string]any)))
			}
		}
	}
	return info
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
			t.Info = NewInfo(value.(map[string]any))
		case "announce":
			t.Announce = value.(string)
		case "announce-list":
			for _, url_list := range value.([]any) {
				for _, url := range url_list.([]any) {
					t.AnnounceList = append(t.AnnounceList, url.(string))
				}

			}
		case "comment":
			t.Comment = value.(string)
		case "created by":
			t.CreatedBy = value.(string)
		case "creation date":
			t.CreationDate = uint64(value.(int))
		case "encoding":
			t.Encoding = value.(string)
		}
	}

	return t, nil
}

func NewTorrent(filename string) (*Torrent, error) {
	file_data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	return NewTorrentFromBencode(file_data)
}
