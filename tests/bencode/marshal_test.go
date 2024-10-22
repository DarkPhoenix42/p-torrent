package bencode_test

import (
	"testing"

	"github.com/DarkPhoenix42/p-torrent/pkg/bencode"
	"github.com/stretchr/testify/assert"
)

var (
	bytesInt64TestData any = map[string]interface{}{
		"announce": []byte("udp://tracker.publicbt.com:80/announce"),
		"announce-list": []interface{}{
			[]interface{}{[]byte("udp://tracker.publicbt.com:80/announce")},
			[]interface{}{[]byte("udp://tracker.openbittorrent.com:80/announce")},
		},
		"comment": []byte("Debian CD from cdimage.debian.org"),
		"info": map[string]interface{}{
			"name":         []byte("debian-8.8.0-arm64-netinst.iso"),
			"length":       170917888,
			"piece length": 262144,
		},
	}

	unmarshalTestData string = "d8:announce38:udp://tracker.publicbt.com:80/announce13:announce-listll38:udp://tracker.publicbt.com:80/announceel44:udp://tracker.openbittorrent.com:80/announceee7:comment33:Debian CD from cdimage.debian.org4:infod6:lengthi170917888e4:name30:debian-8.8.0-arm64-netinst.iso12:piece lengthi262144eee"
	marshalTests             = []struct {
		in  interface{}
		out string
	}{
		{1, "i1e"},
		{42, "i42e"},
		{0, "i0e"},
		{-1, "i-1e"},

		{"", "0:"},
		{"foo", "3:foo"},
		{"bar", "3:bar"},
		{"whatever", "8:whatever"},

		{[]any{}, "le"},
		{[]any{1, 2, 3}, "li1ei2ei3ee"},
		{[]any{4, 5, 6}, "li4ei5ei6ee"},
		{[]any{"foo", "bar"}, "l3:foo3:bare"},
		{[]any{"baz", "qux"}, "l3:baz3:quxe"},

		{map[string]any{}, "de"},
		{map[string]any{"string": 5}, "d6:stringi5ee"},
		{map[string]any{"peers": []any{"hi", "hello"}, "aorts": []any{1, 2}}, "d5:aortsli1ei2ee5:peersl2:hi5:helloee"},

		{bytesInt64TestData, unmarshalTestData},
	}
)

func TestMarshal(t *testing.T) {
	for _, tt := range marshalTests {
		t.Run("", func(t *testing.T) {
			out, err := bencode.Marshal(tt.in)
			if err != nil {
				t.Errorf("Marshal(%v) got error: %v", tt.in, err)
			}
			if string(out) != tt.out {
				t.Errorf("Marshal(%v) = %v; want %v", tt.in, string(out), tt.out)
			}
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	var buffer []byte
	var err error
	for n := 0; n < b.N; n++ {
		buffer, err = bencode.Marshal(bytesInt64TestData)
		if err != nil {
			b.Fatal(err)
		}
	}
	assert.Equal(b, string(unmarshalTestData), string(buffer))
}
