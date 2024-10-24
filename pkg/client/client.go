package client

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/DarkPhoenix42/p-torrent/pkg/bencode"
	"github.com/DarkPhoenix42/p-torrent/pkg/peer"
	"github.com/DarkPhoenix42/p-torrent/pkg/torrent"
	"github.com/rs/zerolog"
)

type Client struct {
	Torrent         *torrent.Torrent
	PeerID          [20]byte
	Tracker         string
	Downloaded      int
	DownloadBuffer  []byte
	Left            int
	TrackerInterval time.Duration
	PeerLock        sync.RWMutex
	Peers           []peer.Peer
	Logger          *zerolog.Logger
}

func NewClient(t *torrent.Torrent, logger *zerolog.Logger) *Client {
	client := Client{
		Torrent:        t,
		PeerID:         [20]byte{},
		Peers:          make([]peer.Peer, 0),
		Downloaded:     0,
		DownloadBuffer: make([]byte, 0),
		Left:           t.GetLength(),
		Logger:         logger,
	}

	_, err := rand.Read(client.PeerID[:])
	if err != nil {
		panic(err)
	}

	return &client
}

func (client *Client) buildTrackerAnnounceURL() (string, error) {

	announce_url, err := url.Parse(client.Torrent.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  {string(client.Torrent.InfoHash[:])},
		"peer_id":    {string(client.PeerID[:])},
		"port":       {"6881"},
		"uploaded":   {"0"},
		"downloaded": {strconv.Itoa(client.Downloaded)},
		"compact":    {"1"},
		"left":       {strconv.Itoa(client.Left)},
		"event":      {"started"},
	}

	announce_url.RawQuery = params.Encode()
	return announce_url.String(), nil
}

func (client *Client) UpdatePeers() error {
	announce_url, err := client.buildTrackerAnnounceURL()
	if err != nil {
		return err
	}

	client.Logger.Info().Msgf("Getting peers from: %s", announce_url)
	http_client := &http.Client{Timeout: 15 * time.Second}
	resp, err := http_client.Get(announce_url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	interval, peers, err := parseTrackerResponse(resp.Body)
	if err != nil {
		return err
	}

	client.TrackerInterval = interval

	for _, p := range peers {
		client.Peers = append(client.Peers, peer.NewPeer(p))
	}

	client.Logger.Info().Msgf("%d peers acquired!", len(client.Peers))
	return nil
}

func parseTrackerResponse(resp_body io.Reader) (time.Duration, []net.Addr, error) {

	bencoded_response, err := io.ReadAll(resp_body)
	if err != nil {
		return 0, nil, err
	}

	decoded_body, err := bencode.UnMarshal(bencoded_response)
	if err != nil {
		return 0, nil, err
	}

	interval := decoded_body.(map[string]any)["interval"].(int)
	peers := []byte(decoded_body.(map[string]any)["peers"].(string))

	peers_list := make([]net.Addr, len(peers)/6)

	for i := 0; i < len(peers); i += 6 {
		ip := net.IP(peers[i : i+4])
		port := binary.BigEndian.Uint16(peers[i+4 : i+6])
		peer := &net.TCPAddr{IP: ip, Port: int(port)}
		peers_list[i/6] = peer
	}

	return time.Duration(interval), peers_list, nil
}

func (client *Client) ConnectToPeers() {
	for _, p := range client.Peers {
		go p.Activate(client.Torrent.InfoHash[:], client.PeerID[:])
	}
}
