package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

func PrintJSON(obj interface{}) {
	bytes, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(bytes))
}

func main() {
	infile, err := os.Open("test_data/manjaro-gnome-17.1.9-stable-x86_64.torrent")

	if err != nil {
		panic(err)
	}

	log.Println("Opened torrent file")

	defer infile.Close()

	reader := bufio.NewReader(infile)

	data, err := io.ReadAll(reader)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Read torrent file: %d bytes\n", len(data))

	decoder := NewDecoder(data)
	
	log.Println("Initialized bencode decoder for torrent")

	output, err := decoder.Decode()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Decoded torrent metainfo")

	infohash, err := decoder.InfoHash()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Computed info_hash: %x\n", infohash)

	// PrintJSON(output)

	metainfo, err := BuildMetainfo(output)

	if err != nil {
		log.Fatal(err)
	}



	// fmt.Println(metainfo)

	req, err := BuildTrackerRequest(infohash, 6881, 0, 0, 0)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Built tracker request")

	params := fmt.Sprintf(
		"info_hash=%s&peer_id=%s&port=6881&uploaded=0&downloaded=0&left=0&compact=0	",
		escapeBytes(req.Info_hash),
		escapeBytes(req.Peer_id),
	)

	url := metainfo.Announce + "?" + params

	log.Printf("Announcing to tracker: %s\n", url)

	response, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Tracker HTTP response: %s\n", response.Status)
	log.Printf("Raw tracker response: %v\n", response)

	body, _ := io.ReadAll(response.Body)
	log.Printf("Read tracker response body: %d bytes\n", len(body))
	fmt.Printf("%q\n", body)

	decoder = NewDecoder(body)

	responseBody, err := decoder.Decode()

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Decoded tracker response")

	fmt.Println(responseBody)

	peerBytes := []byte(responseBody["peers"].(string))
	log.Printf("Parsing peers (raw length=%d bytes)\n", len(peerBytes))
	var peers []net.TCPAddr

	for i := 0; i+6 <= len(peerBytes); i += 6 {
		ip := net.IPv4(
			peerBytes[i],
			peerBytes[i+1],
			peerBytes[i+2],
			peerBytes[i+3],
		)

		port := int(binary.BigEndian.Uint16(peerBytes[i+4 : i+6]))
		peers = append(peers, net.TCPAddr{IP: ip, Port: port})
	}

	log.Printf("Parsed %d peers\n", len(peers))

	for _, peer := range peers {
		fmt.Println(peer.String())
	}
}