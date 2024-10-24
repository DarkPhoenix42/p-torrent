package client

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/DarkPhoenix42/p-torrent/pkg/bencode"
	"github.com/DarkPhoenix42/p-torrent/pkg/torrent"
	"github.com/rs/zerolog"
)

type Client struct {
	Torrent         *torrent.Torrent
	PeerID          [20]byte
	Tracker         string
	TrackerInterval time.Duration
	Peers           []net.Addr
	Logger          *zerolog.Logger
}

func NewClient(t *torrent.Torrent, logger *zerolog.Logger) *Client {
	client := Client{
		Torrent: t,
		PeerID:  [20]byte{},
		Peers:   make([]net.Addr, 0),
		Logger:  logger,
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
		"downloaded": {"0"},
		"compact":    {"1"},
		"left":       {strconv.Itoa(client.Torrent.Info.Length)},
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
	client.Peers = peers

	client.Logger.Info().Msg("Peers updated successfully!")
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
