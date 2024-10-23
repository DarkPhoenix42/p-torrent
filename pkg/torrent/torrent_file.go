package torrent

type TorrentFile struct {
	InfoHash     string
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
	Pieces      string
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
			f.Path = append(f.Path, value.([]string)...)
		}
	}
	return f
}

func NewInfo(m map[string]any) Info {
	i := Info{}
	for key, value := range m {
		switch key {
		case "name":
			i.Name = value.(string)
		case "piece length":
			i.PieceLength = value.(int)
		case "pieces":
			i.Pieces = value.(string)
		case "length":
			i.Length = value.(int)
		case "files":
			for _, file := range value.([]map[string]any) {
				i.Files = append(i.Files, NewFile(file))
			}
		}
	}
	return i
}

func NewTorrentFile(m map[string]any) *TorrentFile {
	t := &TorrentFile{}
	for key, value := range m {
		switch key {
		case "info":
			t.Info = NewInfo(value.(map[string]any))
		case "announce":
			t.Announce = value.(string)
		case "announce-list":
			t.AnnounceList = value.([]string)
		case "comment":
			t.Comment = value.(string)
		case "created by":
			t.CreatedBy = value.(string)
		case "creation date":
			t.CreationDate = value.(uint64)
		case "encoding":
			t.Encoding = value.(string)
		}
	}
	return t
}
