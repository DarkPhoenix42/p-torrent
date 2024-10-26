package client

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/DarkPhoenix42/p-torrent/pkg/bencode"
	"github.com/DarkPhoenix42/p-torrent/pkg/peer"
	"github.com/DarkPhoenix42/p-torrent/pkg/piece"
	"github.com/DarkPhoenix42/p-torrent/pkg/torrent"
	"github.com/rs/zerolog"
)

type Client struct {
	Torrent *torrent.Torrent
	PeerID  [20]byte
	Tracker string

	Downloaded     int
	DownloadBuffer []byte

	Left            int
	TrackerInterval time.Duration
	Peers           map[net.Addr]*peer.Peer
	Logger          *zerolog.Logger

	Work    chan *piece.Piece
	Results chan *piece.Piece
}

func NewClient(t *torrent.Torrent, logger *zerolog.Logger) *Client {
	client := Client{
		Torrent: t,

		PeerID: [20]byte{},
		Peers:  make(map[net.Addr]*peer.Peer, 0),

		Downloaded:     0,
		DownloadBuffer: make([]byte, 0),

		Left:    t.GetLength(),
		Logger:  logger,
		Work:    make(chan *piece.Piece, len(t.Info.Pieces)),
		Results: make(chan *piece.Piece, len(t.Info.Pieces)),
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
	http_client := &http.Client{Timeout: 10 * time.Second}
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
		client.Peers[p] = peer.NewPeer(p, client.Work, client.Results, len(client.Torrent.Info.Pieces))
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
	var wg sync.WaitGroup
	var wMutex sync.RWMutex
	inactive_peers := make([]net.Addr, 0)

	for _, p := range client.Peers {
		wg.Add(1)
		go func(peer *peer.Peer) {
			err := peer.Activate(client.Torrent.InfoHash[:], client.PeerID[:])
			if err != nil {
				client.Logger.Error().Msgf("Failed  to activate peer %s: %s", peer.Addr, err)
				wMutex.Lock()
				inactive_peers = append(inactive_peers, peer.Addr)
				wMutex.Unlock()
			} else {
				client.Logger.Info().Msgf("Activated peer: %s", peer.Addr)
			}

			wg.Done()
		}(p)
	}

	wg.Wait()

	for _, p := range inactive_peers {
		delete(client.Peers, p)
	}

	client.Logger.Info().Msgf("Connected to %d peers", len(client.Peers))

}

func (client *Client) StartDownload() {

	err := client.UpdatePeers()
	if err != nil {
		client.Logger.Error().Msgf("Failed to update peers: %s", err)
		return
	}

	client.ConnectToPeers()

	client.Logger.Info().Msg("Adding work to channels")
	for i := 0; i < len(client.Torrent.Info.Pieces); i++ {
		client.Work <- piece.NewPiece(i, client.Torrent.Info.PieceLength, client.Torrent.Info.Pieces[i])
	}

	client.Logger.Info().Msg("Activating peers for downloading..")
	for _, p := range client.Peers {
		go p.StartDownload()
	}

	client.DownloadBuffer = make([]byte, client.Torrent.Info.PieceLength*len(client.Torrent.Info.Pieces))
	client.Logger.Info().Msgf("Download buffer has been initialized with size %d", len(client.DownloadBuffer))

	downloaded := 0
downloadLoop:
	for {
		select {
		case piece := <-client.Results:
			client.Downloaded += piece.Length
			client.Left -= piece.Length
			downloaded += 1
			client.Logger.Info().Msgf("Downloaded piece #%d [%d/%d]", piece.Index, downloaded, len(client.Torrent.Info.Pieces))
			copy(client.DownloadBuffer[piece.Index*client.Torrent.Info.PieceLength:], piece.Data)

			if client.Left <= 0 {
				client.Logger.Info().Msg("Download complete!")
				break downloadLoop
			}
		}
	}

	downloaded_file, err := os.Create(client.Torrent.Info.Name)
	if err != nil {
		client.Logger.Error().Msgf("Failed to create file: %s", err)
		return
	}

	downloaded_file.Write(client.DownloadBuffer)
}
