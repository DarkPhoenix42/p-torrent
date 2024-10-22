package bencode_test

import (
	"reflect"
	"testing"

	"github.com/DarkPhoenix42/p-torrent/pkg/bencode"
)

var (
	marshalledTestData  string = "d8:announce38:udp://tracker.publicbt.com:80/announce13:announce-listll38:udp://tracker.publicbt.com:80/announceel44:udp://tracker.openbittorrent.com:80/announceee7:comment33:Debian CD from cdimage.debian.org4:infod6:lengthi170917888e4:name30:debian-8.8.0-arm64-netinst.iso12:piece lengthi262144eee"
	umarshalledTestData any    = map[string]interface{}{
		"announce": "udp://tracker.publicbt.com:80/announce",
		"announce-list": []interface{}{
			[]interface{}{"udp://tracker.publicbt.com:80/announce"},
			[]interface{}{"udp://tracker.openbittorrent.com:80/announce"},
		},
		"comment": "Debian CD from cdimage.debian.org",
		"info": map[string]interface{}{
			"name":         "debian-8.8.0-arm64-netinst.iso",
			"length":       170917888,
			"piece length": 262144,
		},
	}

	unMarshalTests = []struct {
		in  string
		out any
	}{
		{"i1e", 1},
		{"i42e", 42},
		{"i0e", 0},
		{"i-1e", -1},

		{"0:", ""},
		{"3:foo", "foo"},
		{"3:bar", "bar"},
		{"8:whatever", "whatever"},

		{"le", []any{}},
		{"li1ei2ei3ee", []any{1, 2, 3}},
		{"li4ei5ei6ee", []any{4, 5, 6}},
		{"l3:foo3:bare", []any{"foo", "bar"}},
		{"l3:baz3:quxe", []any{"baz", "qux"}},

		{"de", map[string]any{}},
		{"d6:stringi5ee", map[string]any{"string": 5}},
		{"d5:aortsli1ei2ee5:peersl2:hi5:helloee", map[string]any{"peers": []any{"hi", "hello"}, "aorts": []any{1, 2}}},

		{marshalledTestData, umarshalledTestData},
	}
)

func TestUnMarshal(t *testing.T) {
	for _, tt := range unMarshalTests {
		t.Run("", func(t *testing.T) {
			var test_case []byte = []byte(tt.in)
			out, err := bencode.UnMarshal(&test_case)
			if err != nil {
				t.Errorf("UnMarshal(%v) got error: %v", tt.in, err)
			}
			if reflect.DeepEqual(out, tt.out) == false {
				t.Errorf("UNnarshal(%v) = %v; want %v", tt.in, out, tt.out)
			}
		})
	}
}

func BenchmarkUnMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, tt := range unMarshalTests {
			var test_case []byte = []byte(tt.in)
			_, _ = bencode.UnMarshal(&test_case)
		}
	}
}
